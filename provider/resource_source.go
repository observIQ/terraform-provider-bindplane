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

	tfresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/terraform-provider-bindplane/internal/client"
	"github.com/observiq/terraform-provider-bindplane/internal/parameter"
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

	resource := model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       "Source",
			Metadata: model.Metadata{
				Name: name,
			},
		},
		Spec: map[string]any{
			"type": sourceType,
		},
	}

	rawParams := d.Get("parameters_json").(string)
	if rawParams != "" {
		var err error
		parameters, err := parameter.StringToParameter(rawParams)
		if err != nil {
			return fmt.Errorf("failed to parse 'parameters_json' for source type '%s' with name '%s': %v", sourceType, name, err)
		}
		resource.Spec["parameters"] = parameters
	}

	bindplane := meta.(*client.BindPlane)

	err := tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutCreate)-time.Minute, func() *tfresource.RetryError {
		err := bindplane.Apply(&resource, rollout)
		if err != nil {
			err := fmt.Errorf("failed to apply resource: %v", err)
			if retryableError(err) {
				return tfresource.RetryableError(err)
			}
			return tfresource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("create retries exhausted: %v", err)
	}

	return resourceSourceRead(d, meta)
}

func resourceSourceRead(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	source := &model.Source{}

	err := tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutRead)-time.Minute, func() *tfresource.RetryError {
		var err error
		name := d.Get("name").(string)
		source, err = bindplane.Source(name)
		if err != nil {
			if retryableError(err) {
				return tfresource.RetryableError(err)
			}
			return tfresource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("read retries exhausted: %v", err)
	}

	if source == nil {
		d.SetId("")
		return nil
	}

	name := source.Name()
	version := source.Version()

	if err := d.Set("name", name); err != nil {
		return fmt.Errorf("failed to set resource name: %v", err)
	}

	if err := d.Set("version", version); err != nil {
		return fmt.Errorf("failed to set resource version: %v", err)
	}

	// TODO(jsirianni): Should Terraform be version aware?
	sourceType := strings.Split(source.Spec.Type, ":")[0]
	if err := d.Set("type", sourceType); err != nil {
		return fmt.Errorf("failed to set resource type: %v", err)
	}

	if len(source.Spec.Parameters) > 0 {
		paramStr, err := parameter.ParametersToString(source.Spec.Parameters)
		if err != nil {
			return fmt.Errorf(
				"failed to convert source parameters into 'parameters_json' for source type '%s' with name '%s': %v",
				sourceType, source.Name(), err)
		}

		if err := d.Set("parameters_json", paramStr); err != nil {
			return fmt.Errorf("failed to set resource parameters_json: %v", err)
		}
	}

	d.SetId(source.ID())

	return nil
}

func resourceSourceDelete(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	err := tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutDelete)-time.Minute, func() *tfresource.RetryError {
		name := d.Get("name").(string)
		err := bindplane.DeleteSource(name)
		if err != nil {
			err := fmt.Errorf("failed to delete source '%s' by name: %v", name, err)
			if retryableError(err) {
				return tfresource.RetryableError(err)
			}
			return tfresource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("delete retries exhausted: %v", err)
	}

	return resourceSourceRead(d, meta)
}
