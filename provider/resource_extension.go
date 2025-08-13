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

func resourceExtension() *schema.Resource {
	return &schema.Resource{
		Create: resourceExtensionCreate,
		Update: resourceExtensionCreate,
		Read:   resourceExtensionRead,
		Delete: resourceExtensionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceExtensionImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the extension.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "The extension type to use for extension creation.",
			},
			"parameters_json": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "A JSON object with options used to configure the extension.",
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

func resourceExtensionCreate(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	extensionType := d.Get("type").(string)
	name := d.Get("name").(string)
	rollout := d.Get("rollout").(bool)

	// If id is unset, it means Terraform has not previously created
	// this resource. Check to ensure a resource with this name does
	// not already exist.
	if d.Id() == "" {
		c, err := bindplane.Extension(name)
		if err != nil {
			return err
		}
		if c != nil {
			return fmt.Errorf("extension with name '%s' already exists with id '%s'", name, c.ID())
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

	r, err := resource.AnyResourceV1(id, name, extensionType, model.KindExtension, parameters, nil, nil)
	if err != nil {
		return err
	}

	ctx := context.Background()
	timeout := d.Timeout(schema.TimeoutCreate) - time.Minute
	if err := bindplane.ApplyWithRetry(ctx, timeout, &r, rollout); err != nil {
		return err
	}

	return resourceExtensionRead(d, meta)
}

func resourceExtensionRead(d *schema.ResourceData, meta any) error {
	return genericResourceRead(model.KindExtension, d, meta)
}

func resourceExtensionDelete(d *schema.ResourceData, meta any) error {
	return genericResourceDelete(model.KindExtension, d, meta)
}

func resourceExtensionImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	return genericResourceImport(model.KindExtension, d, meta)
}
