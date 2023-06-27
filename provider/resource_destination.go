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

	parameters, err := parameter.StringToParameter(d.Get("parameters_json").(string))
	if err != nil {
		return fmt.Errorf("failed to parse 'parameters_json' for destination type '%s' with name '%s': %v", destType, name, err)
	}

	resource := model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1alpha",
			Kind:       "Destination",
			Metadata: model.Metadata{
				Name: name,
			},
		},
		Spec: map[string]any{
			"type":       destType,
			"parameters": parameters,
		},
	}

	bindplane := meta.(*client.BindPlane)

	id := ""
	err = tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutCreate)-time.Minute, func() *tfresource.RetryError {
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

	// Split the destination type and version
	// googlecloud:1 --> googlecloud, 1.
	// TODO(jsirianni): Is this safe? Should versioned resources be handled differently?
	// TODO(jsirianni): Can we just omit updating the destinationt type field? Probably not. I suppose
	// someone could delete the destinaion and re-create it by name but then its id would change, causing
	// it to be re-created by tf?
	destinationType := strings.Split(destination.Spec.Type, ":")[0]
	if err := d.Set("type", destinationType); err != nil {
		return fmt.Errorf("failed to set resource type: %v", err)
	}

	paramStr, err := parameter.ParametersToString(destination.Spec.Parameters)
	if err != nil {
		return fmt.Errorf(
			"failed to convert destination parameters into 'parameters_json' for destination type '%s' with name '%s': %v",
			destinationType, destination.Name(), err)
	}
	d.Set("parameters_json", paramStr)

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
