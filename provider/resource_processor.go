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

func resourceProcessor() *schema.Resource {
	return &schema.Resource{
		Create: resourceProcessorCreate,
		Update: resourceProcessorCreate,
		Read:   resourceProcessorRead,
		Delete: resourceProcessorDelete,
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

	resource := model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       "Processor",
			Metadata: model.Metadata{
				Name: name,
			},
		},
		Spec: map[string]any{
			"type":       processorType,
			"parameters": parameters,
		},
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

	return resourceProcessorRead(d, meta)
}

func resourceProcessorRead(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	processor := &model.Processor{}

	err := tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutRead)-time.Minute, func() *tfresource.RetryError {
		var err error
		name := d.Get("name").(string)
		processor, err = bindplane.Processor(name)
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

	if processor == nil {
		d.SetId("")
		return nil
	}

	name := processor.Name()
	version := processor.Version()

	if err := d.Set("name", name); err != nil {
		return fmt.Errorf("failed to set resource name: %v", err)
	}

	if err := d.Set("version", version); err != nil {
		return fmt.Errorf("failed to set resource version: %v", err)
	}

	// TODO(jsirianni): Should Terraform be version aware?
	processorType := strings.Split(processor.Spec.Type, ":")[0]
	if err := d.Set("type", processorType); err != nil {
		return fmt.Errorf("failed to set resource type: %v", err)
	}

	paramStr, err := parameter.ParametersToString(processor.Spec.Parameters)
	if err != nil {
		return err
	}

	if err := d.Set("parameters_json", string(paramStr)); err != nil {
		return fmt.Errorf("failed to set resource parameters_json: %v", err)
	}

	d.SetId(processor.ID())

	return nil
}

func resourceProcessorDelete(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	err := tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutDelete)-time.Minute, func() *tfresource.RetryError {
		name := d.Get("name").(string)
		err := bindplane.DeleteProcessor(name)
		if err != nil {
			err := fmt.Errorf("failed to delete processor '%s' by name: %v", name, err)
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

	return resourceProcessorRead(d, meta)
}
