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

	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/terraform-provider-bindplane/internal/client"
	"github.com/observiq/terraform-provider-bindplane/internal/configuration"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"

	tfresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceConfigurationCreate,
		Update: resourceConfigurationCreate, // Run create as update
		Read:   resourceConfigurationRead,
		Delete: resourceGenericConfigurationDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Required: true,
				ForceNew: false,
			},
			"match_labels": {
				Type:     schema.TypeMap,
				Required: true,
				ForceNew: false,
			},
			"sources": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"destinations": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

func resourceConfigurationCreate(d *schema.ResourceData, meta any) error {
	name := d.Get("name").(string)
	rollout := d.Get("rollout").(bool)

	labels, err := stringMapFromTFMap(d.Get("labels").(map[string]any))
	if err != nil {
		return fmt.Errorf("failed to read labels from resource configuration: %v", err)
	}

	matchLabels, err := stringMapFromTFMap(d.Get("match_labels").(map[string]any))
	if err != nil {
		return fmt.Errorf("failed to read match labels from resource configuration: %v", err)
	}

	// Build list of source names
	var sources []string
	if v := d.Get("sources").(*schema.Set); v != nil {
		for _, v := range v.List() {
			name := v.(string)
			sources = append(sources, name)
		}
	}

	// Build list of destination names
	var destinations []string
	if v := d.Get("destinations").(*schema.Set); v != nil {
		for _, v := range v.List() {
			name := v.(string)
			destinations = append(destinations, name)
		}
	}

	opts := []configuration.Option{
		configuration.WithName(name),
		configuration.WithLabels(labels),
		configuration.WithMatchLabels(matchLabels),
		configuration.WithSourcesByName(sources),
		configuration.WithDestinationsByName(destinations),
	}

	config, err := configuration.NewV1(opts...)
	if err != nil {
		return fmt.Errorf("failed to create new configuration: %v", err)
	}

	resource := resource.AnyResourceFromConfiguration(config)

	bindplane := meta.(*client.BindPlane)

	err = tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutCreate)-time.Minute, func() *tfresource.RetryError {
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

	return resourceConfigurationRead(d, meta)
}

func resourceConfigurationRead(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	config := &model.Configuration{}

	err := tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutRead)-time.Minute, func() *tfresource.RetryError {
		var err error
		name := d.Get("name").(string)
		config, err = bindplane.Configuration(name)
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

	matchLabels := make(map[string]string)
	for k, v := range config.Spec.Selector.MatchLabels {
		matchLabels[k] = v
	}
	if err := d.Set("match_labels", matchLabels); err != nil {
		return fmt.Errorf("failed to set resource match labels: %v", err)
	}

	// for _, source := range config.Spec.Sources {
	// 	paramStr, err := parameter.ParametersToString(source.Parameters)
	// 	if err != nil {
	// 		return fmt.Errorf(
	// 			"failed to convert source parameters into 'parameters_json' for source type '%s': %v",
	// 			source.Type, err)
	// 	}
	// 	if err := d.Set("parameters_json", paramStr); err != nil {
	// 		return fmt.Errorf("failed to set resource parameters: %v", err)
	// 	}
	// }

	d.SetId(config.ID())
	return nil
}
