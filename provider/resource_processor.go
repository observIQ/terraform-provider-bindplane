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
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/observiq/terraform-provider-bindplane/internal/parameter"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"
)

func resourceProcessor() *schema.Resource {
	return &schema.Resource{
		Create: resourceProcessorCreate,
		Update: resourceProcessorCreate,
		Read:   resourceProcessorRead,
		Delete: resourceProcessorDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the processor.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "The destination type to use for processor creation.",
			},
			"parameters_json": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "A JSON object with options used to configure the processor.",
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

func resourceProcessorCreate(d *schema.ResourceData, meta any) error {
	processorType := d.Get("type").(string)
	name := d.Get("name").(string)
	rollout := d.Get("rollout").(bool)

	parameters := []model.Parameter{}
	if s := d.Get("parameters_json").(string); s != "" {
		params, err := parameter.StringToParameter(s)
		if err != nil {
			return err
		}
		parameters = params
	}

	r, err := resource.AnyResourceV1(name, processorType, model.KindProcessor, parameters)
	if err != nil {
		return err
	}

	bindplane := meta.(*client.BindPlane)
	ctx := context.TODO()
	timeout := d.Timeout(schema.TimeoutCreate) - time.Minute
	if err := bindplane.ApplyWithRetry(ctx, timeout, &r, rollout); err != nil {
		return err
	}

	return resourceProcessorRead(d, meta)
}

func resourceProcessorRead(d *schema.ResourceData, meta any) error {
	return genericResourceRead(model.KindProcessor, d, meta)
}

func resourceProcessorDelete(d *schema.ResourceData, meta any) error {
	return genericResourceDelete(model.KindProcessor, d, meta)
}
