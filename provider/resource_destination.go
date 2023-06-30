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
	"github.com/observiq/terraform-provider-bindplane/internal/client"
	"github.com/observiq/terraform-provider-bindplane/internal/parameter"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"
)

func resourceDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceDestinationCreate,
		Update: resourceDestinationCreate,
		Read:   resourceDestinationRead,
		Delete: resourceDestinationDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the destination.",
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
				ForceNew: false,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"parameters_json": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"rollout": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: false,
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
	destType := d.Get("type").(string)
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

	r, err := resource.AnyResourceV1(name, destType, model.KindDestination, parameters)
	if err != nil {
		return err
	}

	bindplane := meta.(*client.BindPlane)
	ctx := context.TODO()
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
