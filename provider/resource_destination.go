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

	tfresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/terraform-provider-bindplane/internal/client"
)

func resourceDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceDestinationCreate,
		Update: resourceDestinationCreate,
		Read:   resourceDestinationRead,
		Delete: resourceDestinationDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// Destination type, such as `googlecloud` or `logging`.
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			// Key value pairs used to configure the destination. Keys must
			// match the destinaton type's parameters.
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
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
	paramList := []model.Parameter{}
	params := d.Get("parameters").(map[string]any)
	for k, v := range params {
		p := model.Parameter{
			Name:  k,
			Value: v,
		}
		paramList = append(paramList, p)
	}

	resource := model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1alpha",
			Kind:       "Destination",
			Metadata: model.Metadata{
				Name: d.Get("name").(string),
			},
		},
		Spec: map[string]any{
			"type":       d.Get("type").(string),
			"parameters": paramList,
		},
	}

	bindplane := meta.(*client.BindPlane)

	id := ""
	err := tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutCreate)-time.Minute, func() *tfresource.RetryError {
		var err error
		id, err = bindplane.Apply(&resource)
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
	d.SetId(id) // TODO: is this necessary or will read handle this?

	return resourceDestinationRead(d, meta)
}

func resourceDestinationRead(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	destination := &model.Destination{}

	err := tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutRead)-time.Minute, func() *tfresource.RetryError {
		var err error
		name := d.Get("name").(string)
		destination, err = bindplane.Destination(name)
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

	if destination == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("name", destination.Name()); err != nil {
		return fmt.Errorf("failed to set resource name: %v", err)
	}

	if err := d.Set("type", destination.Spec.Type); err != nil {
		return fmt.Errorf("failed to set resource type: %v", err)
	}

	params := map[string]any{}
	for _, param := range destination.Spec.Parameters {
		params[param.Name] = param.Value
	}

	d.SetId(destination.ID())

	return nil
}

func resourceDestinationDelete(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	err := tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutDelete)-time.Minute, func() *tfresource.RetryError {
		name := d.Get("name").(string)
		err := bindplane.DeleteDestination(name)
		if err != nil {
			err := fmt.Errorf("failed to delete destination '%s' by name: %v", name, err)
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

	return resourceDestinationRead(d, meta)
}
