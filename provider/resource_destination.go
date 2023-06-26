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

func resourceDestinationGoogleCloud() *schema.Resource {
	return &schema.Resource{
		Create: resourceDestinationGoogleCloudCreate,
		Update: resourceDestinationGoogleCloudCreate,
		Read:   resourceDestinationGoogleCloudRead,
		Delete: resourceDestinationGoogleCloudDelete,
		Schema: map[string]*schema.Schema{
			// Metadata
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
			},
			// parameters
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"auth_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "auto",
			},
			"credentials": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "",
			},
			"credentials_file": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(maxTimeout),
			Read:   schema.DefaultTimeout(maxTimeout),
			Delete: schema.DefaultTimeout(maxTimeout),
		},
	}
}

func resourceDestinationGoogleCloudCreate(d *schema.ResourceData, meta any) error {
	l, err := stringMapFromTFMap(d.Get("labels").(map[string]any))
	if err != nil {
		return fmt.Errorf("failed to read labels from resource configuration: %v", err)
	}
	labels, err := model.LabelsFromMap(l)
	if err != nil {
		return fmt.Errorf("failed to set destination labels: %v", err)
	}

	resource := model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1alpha",
			Kind:       "Destination",
			Metadata: model.Metadata{
				ID:     d.Get("id").(string),
				Name:   d.Get("name").(string),
				Labels: labels,
			},
		},
		Spec: map[string]any{
			"type": "googlecloud",
			"parameters": []map[string]any{
				{
					"name":  "project",
					"value": d.Get("project").(string),
				},
				{
					"name":  "auth_type",
					"value": d.Get("auth_type").(string),
				},
				{
					"name":  "credentials",
					"value": d.Get("credentials").(string),
				},
				{
					"name":  "credentials_file",
					"value": d.Get("credentials_file").(string),
				},
			},
		},
	}

	bindplane := meta.(*client.BindPlane)

	id := ""
	err = tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutCreate)-time.Minute, func() *tfresource.RetryError {
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

	return resourceDestinationGoogleCloudRead(d, meta)
}

func resourceDestinationGoogleCloudRead(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	config := &model.Destination{}

	err := tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutRead)-time.Minute, func() *tfresource.RetryError {
		var err error
		name := d.Get("name").(string)
		config, err = bindplane.Destination(name)
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

	if config == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("name", config.Name()); err != nil {
		return fmt.Errorf("failed to set resource name: %v", err)
	}

	if err := d.Set("labels", config.Metadata.Labels.AsMap()); err != nil {
		return fmt.Errorf("failed to set resource labels: %v", err)
	}

	// TODO(jsirianni): Not safe to assume spec param values
	// are always a string. They could change if the underlying source type changes,
	// but it is unlikely. We should handle this gracefully.

	for _, p := range config.Spec.Parameters {
		if p.Name == "project" {
			if err := d.Set("project", p.Value.(string)); err != nil {
				return fmt.Errorf("failed to set resource project: %v", err)
			}
		}
	}

	for _, p := range config.Spec.Parameters {
		if p.Name == "auth_type" {
			if err := d.Set("auth_type", p.Value.(string)); err != nil {
				return fmt.Errorf("failed to set resource auth_type: %v", err)
			}
		}
	}

	for _, p := range config.Spec.Parameters {
		if p.Name == "credentials" {
			if err := d.Set("credentials", p.Value.(string)); err != nil {
				return fmt.Errorf("failed to set resource credentials: %v", err)
			}
		}
	}

	for _, p := range config.Spec.Parameters {
		if p.Name == "credentials_file" {
			if err := d.Set("credentials_file", p.Value.(string)); err != nil {
				return fmt.Errorf("failed to set resource credentials_file: %v", err)
			}
		}
	}

	d.SetId(config.ID())

	return nil
}

func resourceDestinationGoogleCloudDelete(d *schema.ResourceData, meta any) error {
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

	return resourceDestinationGoogleCloudRead(d, meta)
}
