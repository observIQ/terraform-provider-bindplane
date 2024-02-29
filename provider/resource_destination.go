// Copyright  observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/observiq/terraform-provider-bindplane/internal/parameter"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"
)

func resourceDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceDestinationCreate,
		Update: resourceDestinationCreate,
		Read:   resourceDestinationRead,
		Delete: resourceDestinationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDestinationImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the destination.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "The destination type to use for destination creation.",
			},
			"parameters_json": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "A JSON object with options used to configure the destination.",
			},
			"rollout": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "Whether or not to trigger a rollout automatically when a configuration is updated. When set to true, BindPlane OP will automatically roll out the configuration change to managed agents.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(maxTimeout),
			Read:   schema.DefaultTimeout(maxTimeout),
			Delete: schema.DefaultTimeout(maxTimeout),
		},
	}
}

func resourceDestinationCreate(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	destType := d.Get("type").(string)
	name := d.Get("name").(string)
	rollout := d.Get("rollout").(bool)

	// If id is unset, it means Terraform has not previously created
	// this resource. Check to ensure a resource with this name does
	// not already exist.
	if d.Id() == "" {
		c, err := bindplane.Destination(name)
		if err != nil {
			return err
		}
		if c != nil {
			return fmt.Errorf("destination with name '%s' already exists with id '%s'", name, c.ID())
		}
	}

	parameters := []model.Parameter{}
	if s := d.Get("parameters_json").(string); s != "" {
		params, err := parameter.StringToParameter(s)
		if err != nil {
			return err
		}
		parameters = params
	}

	r, err := resource.AnyResourceV1(name, destType, model.KindDestination, parameters)
	if err != nil {
		return err
	}

	ctx := context.Background()
	timeout := d.Timeout(schema.TimeoutCreate) - time.Minute
	if err := bindplane.ApplyWithRetry(ctx, timeout, &r, rollout); err != nil {
		return err
	}

	return resourceDestinationRead(d, meta)
}

func resourceDestinationRead(d *schema.ResourceData, meta any) error {
	return genericResourceRead(model.KindDestination, d, meta)
}

func resourceDestinationDelete(d *schema.ResourceData, meta any) error {
	return genericResourceDelete(model.KindDestination, d, meta)
}

func resourceDestinationImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	bindplane := meta.(*client.BindPlane)

	// When importing, name is not set in the state so we need to grab
	// the ID instead, which is the same as "name".
	destName := d.Id()

	g, err := bindplane.GenericResource(model.KindDestination, destName)
	if err != nil {
		return nil, err
	}

	// bindplane.GenericResource will return a nil error if the resource
	// does not exist. It is up to the caller to check.
	if g == nil {
		return nil, fmt.Errorf("destination with name '%s' does not exist", destName)
	}

	// Add the name to state, which will cause the import to succeed.
	d.Set("name", g.Name)

	return []*schema.ResourceData{d}, nil
}
