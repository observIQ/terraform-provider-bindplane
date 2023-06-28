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
	"encoding/json"
	"errors"
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
			"platform": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(val any, _ string) (warns []string, errs []error) {
					platform := val.(string)
					if !isValidPlatform(platform) {
						errs = append(errs, fmt.Errorf("%s is not a valid platform", platform))
					}
					return
				},
			},
			"labels": {
				Type:     schema.TypeMap,
				Required: true,
				ForceNew: false,
				ValidateFunc: func(val any, _ string) (warns []string, errs []error) {
					labels := val.(map[string]any)
					_, ok := labels["platform"]
					if ok {
						errs = append(errs, errors.New("label 'platform' will be overwritten by the configured platform"))
					}
					return
				},
			},
			"match_labels": {
				Type:     schema.TypeMap,
				Computed: true,
				ForceNew: false,
			},
			"source": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
				},
			},
			"sources": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"destinations": {
				Type:     schema.TypeSet,
				Optional: true,
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
	labels["platform"] = d.Get("platform").(string)

	// Match labels should always be configuration=<name>
	matchLabels := map[string]string{
		"configuration": name,
	}

	// // Build inline sources
	inlineSources := []model.ResourceConfiguration{}
	if d.Get("source") != nil {
		inlineSourcesRaw := d.Get("source").([]any)

		for _, v := range inlineSourcesRaw {
			inlineSource := v.(map[string]any)

			// Build source without params
			source := model.ResourceConfiguration{
				ParameterizedSpec: model.ParameterizedSpec{
					Type: inlineSource["type"].(string),
				},
			}

			if paramStr := inlineSource["parameters_json"].(string); paramStr != "" {
				params := []model.Parameter{}
				if err := json.Unmarshal([]byte(paramStr), &params); err != nil {
					return fmt.Errorf("failed to unmarshal parameters '%s': %v", paramStr, err)
				}
				source.ParameterizedSpec.Parameters = params
			}

			inlineSources = append(inlineSources, source)
		}
	}

	// Build list of source names
	// TODO(jsirianni): Ensure this still works when no sources
	// are configured.
	var sources []string
	if v := d.Get("sources").(*schema.Set); v != nil {
		for _, v := range v.List() {
			name := v.(string)
			sources = append(sources, name)
		}
	}

	// Build list of destination names
	// TODO(jsirianni): Ensure this still works when no destinations
	// are configured.
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
		configuration.WithSourcesInline(inlineSources),
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

	labels := config.Metadata.Labels.AsMap()
	platform, ok := labels["platform"]
	if ok {
		if err := d.Set("platform", platform); err != nil {
			return fmt.Errorf("failed to set resource platform: %v", err)
		}
		// Remove the platform label from the labels map
		// because Terraform's state does not expect it.
		delete(labels, "platform")
	}

	// Save the labels map to state, which has the 'platform' label removed.
	if err := d.Set("labels", labels); err != nil {
		return fmt.Errorf("failed to set resource labels: %v", err)
	}

	matchLabels := make(map[string]string)
	for k, v := range config.Spec.Selector.MatchLabels {
		matchLabels[k] = v
	}
	if err := d.Set("match_labels", matchLabels); err != nil {
		return fmt.Errorf("failed to set resource match labels: %v", err)
	}

	// TODO(jsirianni): Read source params and save to state.
	// Right now we do not have a way to identify embeded sources, therefor
	// we will run into issues when it comes to mutliple sources of the same
	// source type.
	// Embeded sources
	/*for _, source := range config.Spec.Sources {

	}*/

	d.SetId(config.ID())
	return nil
}
