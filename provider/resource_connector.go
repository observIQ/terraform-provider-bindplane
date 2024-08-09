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
	"github.com/observiq/terraform-provider-bindplane/internal/component"
	"github.com/observiq/terraform-provider-bindplane/internal/parameter"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"
)

func resourceConnector() *schema.Resource {
	return &schema.Resource{
		Create: resourceConnectorCreate,
		Update: resourceConnectorCreate,
		Read:   resourceConnectorRead,
		Delete: resourceConnectorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceConnectorImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the connector.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "The connector type to use for connector creation.",
			},
			"parameters_json": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				Description:      "A JSON object with options used to configure the connector.",
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
			},
			"rollout": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "Whether or not to trigger a rollout automatically when a configuration is updated. When set to true, Bindplane will automatically roll out the configuration change to managed agents.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(maxTimeout),
			Read:   schema.DefaultTimeout(maxTimeout),
			Delete: schema.DefaultTimeout(maxTimeout),
		},
	}
}

func resourceConnectorCreate(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	connectorType := d.Get("type").(string)
	name := d.Get("name").(string)
	rollout := d.Get("rollout").(bool)

	// If id is unset, it means Terraform has not previously created
	// this resource. Check to ensure a resource with this name does
	// not already exist.
	if d.Id() == "" {
		c, err := bindplane.Connector(name)
		if err != nil {
			return err
		}
		if c != nil {
			return fmt.Errorf("connector with name '%s' already exists with id '%s'", name, c.ID())
		}

		// If a source does not already exist with this name
		// and an ID is not set, generate and ID.
		d.SetId(component.NewResourceID())
	}

	id := d.Id()

	parameters := []model.Parameter{}
	if s := d.Get("parameters_json").(string); s != "" {
		params, err := parameter.StringToParameter(s)
		if err != nil {
			return err
		}
		parameters = params
	}

	displayName := ""
	if v := d.Get("display_name").(string); v != "" {
		displayName = v
	}

	description := ""
	if v := d.Get("description").(string); v != "" {
		description = v
	}

	r, err := resource.AnyResourceV1(id, name, connectorType, model.KindConnector, parameters, nil, displayName, description)
	if err != nil {
		return err
	}

	ctx := context.Background()
	timeout := d.Timeout(schema.TimeoutCreate) - time.Minute
	if err := bindplane.ApplyWithRetry(ctx, timeout, &r, rollout); err != nil {
		return err
	}

	return resourceConnectorRead(d, meta)
}

func resourceConnectorRead(d *schema.ResourceData, meta any) error {
	return genericResourceRead(model.KindConnector, d, meta)
}

func resourceConnectorDelete(d *schema.ResourceData, meta any) error {
	return genericResourceDelete(model.KindConnector, d, meta)
}

func resourceConnectorImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	return genericResourceImport(model.KindConnector, d, meta)
}
