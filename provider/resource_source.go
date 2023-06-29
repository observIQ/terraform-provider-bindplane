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
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/terraform-provider-bindplane/internal/client"
	"github.com/observiq/terraform-provider-bindplane/internal/parameter"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"
)

// TODO(jsirianni): Decide if sources should be supported. Currently not implemented by the provider.
func resourceSource() *schema.Resource {
	return &schema.Resource{
		Create: resourceSourceCreate,
		Update: resourceSourceCreate,
		Read:   resourceSourceRead,
		Delete: resourceSourceDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

func resourceSourceCreate(d *schema.ResourceData, meta any) error {
	sourceType := d.Get("type").(string)
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

	r, err := resource.AnyResourceV1(name, sourceType, model.KindSource, parameters)
	if err != nil {
		return err
	}

	bindplane := meta.(*client.BindPlane)
	ctx := context.TODO()
	timeout := d.Timeout(schema.TimeoutCreate) - time.Minute
	if err := bindplane.ApplyWithRetry(ctx, timeout, &r, rollout); err != nil {
		return err
	}

	return resourceSourceRead(d, meta)
}

func resourceSourceRead(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	source, err := bindplane.Source(d.Get("name").(string))
	if err != nil {
		return err
	}

	if source == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("name", source.Name()); err != nil {
		return fmt.Errorf("failed to set resource name: %v", err)
	}

	if err := d.Set("version", source.Version()); err != nil {
		return fmt.Errorf("failed to set resource version: %v", err)
	}

	// TODO(jsirianni): Should Terraform be version aware?
	sourceType := strings.Split(source.Spec.Type, ":")[0]
	if err := d.Set("type", sourceType); err != nil {
		return fmt.Errorf("failed to set resource type: %v", err)
	}

	paramStr, err := parameter.ParametersToString(source.Spec.Parameters)
	if err != nil {
		return err
	}

	if err := d.Set("parameters_json", string(paramStr)); err != nil {
		return fmt.Errorf("failed to set resource parameters_json: %v", err)
	}

	d.SetId(source.ID())

	return nil
}

func resourceSourceDelete(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)
	err := bindplane.DeleteSource(d.Get("name").(string))
	if err != nil {
		return err
	}
	return resourceSourceRead(d, meta)
}
