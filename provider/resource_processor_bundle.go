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
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/observiq/terraform-provider-bindplane/internal/component"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"
)

func resourceProcessorBundle() *schema.Resource {
	return &schema.Resource{
		Create: resourceProcessorBundleCreate,
		Update: resourceProcessorBundleCreate,
		Read:   resourceProcessorBundleRead,
		Delete: resourceProcessorBundleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceProcessorBundleImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the processor bundle.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    false,
				Description: "The type of the processor bundle.",
			},
			"parameters_json": {
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    false,
				Description: "Not implemented.",
			},
			"processor": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The processors to use for the processor bundle.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the processor.",
						},
					},
				},
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

func resourceProcessorBundleCreate(d *schema.ResourceData, meta any) error {

	bindplane := meta.(*client.BindPlane)

	processorType := d.Get("type").(string)
	if processorType == "" {
		// TODO(jsirianni): Move to default const
		processorType = "processor_bundle"
	}

	name := d.Get("name").(string)
	rollout := d.Get("rollout").(bool)

	// If id is unset, it means Terraform has not previously created
	// this resource. Check to ensure a resource with this name does
	// not already exist.
	if d.Id() == "" {
		c, err := bindplane.Processor(name)
		if err != nil {
			return err
		}
		if c != nil {
			return fmt.Errorf("processor with name '%s' already exists with id '%s'", name, c.ID())
		}

		// If a source does not already exist with this name
		// and an ID is not set, generate and ID.
		d.SetId(component.NewResourceID())
	}

	id := d.Id()

	// Using resource configuration instead of []string (names)
	// to allow for future use of type + parameters_json.
	processors := []model.ResourceConfiguration{}
	if d.Get("processor") != nil {
		processorsRaw := d.Get("processor").([]any)
		for _, v := range processorsRaw {
			processorRaw := v.(map[string]any)

			processor := model.ResourceConfiguration{}
			if processorRaw["name"] != nil {
				processor.Name = processorRaw["name"].(string)

				// If name was set, return early to avoid
				// setting type and parameters
				processors = append(processors, processor)
				continue
			}

			processors = append(processors, processor)
		}
	}

	r, err := resource.AnyResourceV1(id, name, processorType, model.KindProcessor, nil, processors)
	if err != nil {
		return err
	}

	ctx := context.Background()
	timeout := d.Timeout(schema.TimeoutCreate) - time.Minute
	if err := bindplane.ApplyWithRetry(ctx, timeout, &r, rollout); err != nil {
		return err
	}

	return resourceProcessorBundleRead(d, meta)
}

func resourceProcessorBundleRead(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)
	resourceName := d.Get("name").(string)

	g, err := bindplane.GenericResource(model.KindProcessor, resourceName)
	if err != nil {
		return err
	}

	// A nil return from GenericResource indicates that the resource
	// did not exist. Terraform read operations should always set the
	// ID to "" and return a nil error. This will allow Terraform to
	// re-create the resource or comfirm that it was deleted.
	if g == nil {
		d.SetId("")
		return nil
	}

	// Save values returned by bindplane to Terraform's state

	d.SetId(g.ID)

	// If the state ID is set but differs from the ID returned by,
	// bindplane, mark the resource to be re-created by unsetting
	// the ID. This will cause Terraform to attempt to create the resource
	// instead of updating it. The creation step will fail because
	// the resource already exists. This behavior is desirable, it will
	// prevent Terraform from modifying resources created by other means.
	if id := d.Id(); id != "" {
		if g.ID != d.Id() {
			d.SetId("")
			return nil
		}
	}

	if err := d.Set("name", g.Name); err != nil {
		return err
	}

	rType := strings.Split(g.Spec.Type, ":")[0]
	if err := d.Set("type", rType); err != nil {
		return err
	}

	processorBlocks := []map[string]any{}
	for _, p := range g.Spec.Processors {
		processor := map[string]any{}
		processor["name"] = strings.Split(p.Name, ":")[0]
		processorBlocks = append(processorBlocks, processor)
	}
	return d.Set("processor", processorBlocks)
}

func resourceProcessorBundleDelete(d *schema.ResourceData, meta any) error {
	return genericResourceDelete(model.KindProcessor, d, meta)
}

func resourceProcessorBundleImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	return genericResourceImport(model.KindProcessor, d, meta)
}
