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

// This error generally happens only when BindPlane is being provisioned for the
// first time, or if the network connection is flaky. This is a retryable error.
const (
	errClientConnectionRefused = "connect: connection refused"
	errClientTimeoutRetry      = "Client.Timeout exceeded while awaiting headers"
)

func resourceConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceConfigurationCreate,
		Update: resourceConfigurationCreate, // Run create as update
		Read:   resourceConfigurationRead,
		Delete: resourceConfigurationDelete,
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
						"parameters": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: false,
						},
					},
				},
			},
			"destinations": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// TODO(jsirianni): Remove or make sure it is safe to keep.
			"raw_configuration": {
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

func resourceConfigurationCreate(d *schema.ResourceData, meta any) error {
	labels, err := stringMapFromTFMap(d.Get("labels").(map[string]any))
	if err != nil {
		return fmt.Errorf("failed to read labels from resource configuration: %v", err)
	}

	matchLabels, err := stringMapFromTFMap(d.Get("match_labels").(map[string]any))
	if err != nil {
		return fmt.Errorf("failed to read match labels from resource configuration: %v", err)
	}

	opts := []configuration.Option{
		configuration.WithName(d.Get("name").(string)),
		configuration.WithLabels(labels),
		configuration.WithMatchLabels(matchLabels),
	}

	if raw, ok := d.Get("raw_configuration").(string); ok && raw != "" {
		opts = append(opts, configuration.WithRawOTELConfig(raw))
	} else {
		var sources []model.ResourceConfiguration

		// raw list of sources
		if v := d.Get("source").([]any); v != nil {
			sources = make([]model.ResourceConfiguration, len(v))

			// for each raw source
			for _, raw := range v {
				data := raw.(map[string]interface{})

				sourceType, ok := data["type"].(string)
				if !ok || sourceType == "" {
					return errors.New("source configuration's 'type' parameter must be set")
				}

				params := []model.Parameter{}
				rawParams, ok := data["parameters"].([]map[string]string)
				if ok {
					for _, p := range rawParams {
						param := model.Parameter{
							Name:  p["name"],
							Value: p["value"],
						}
						params = append(params, param)
					}
				}

				source := model.ResourceConfiguration{
					ParameterizedSpec: model.ParameterizedSpec{
						Type:       sourceType,
						Parameters: params,
					},
				}
				sources = append(sources, source)
			}
		}
		opts = append(opts, configuration.WithSources(sources))

		var destinations []model.ResourceConfiguration
		if v := d.Get("destinations").(*schema.Set); v != nil {
			destinations = make([]model.ResourceConfiguration, v.Len())
			for _, v := range v.List() {
				d := model.ResourceConfiguration{
					Name: v.(string),
				}
				destinations = append(destinations, d)
			}
		}
		opts = append(opts, configuration.WithDestinations(destinations))
	}

	config, err := configuration.NewV1Alpha(opts...)
	if err != nil {
		return fmt.Errorf("failed to create new v1alpha configuration: %v", err)
	}

	resource := resource.AnyResourceFromConfiguration(config)

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

	if err := d.Set("raw_configuration", config.Spec.Raw); err != nil {
		return fmt.Errorf("failed to set resource raw_configuration: %v", err)
	}

	matchLabels := make(map[string]string)
	for k, v := range config.Spec.Selector.MatchLabels {
		matchLabels[k] = v
	}
	if err := d.Set("match_labels", matchLabels); err != nil {
		return fmt.Errorf("failed to set resource match labels: %v", err)
	}

	d.SetId(config.ID())
	return nil
}

func resourceConfigurationDelete(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	err := tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutDelete)-time.Minute, func() *tfresource.RetryError {
		name := d.Get("name").(string)
		err := bindplane.DeleteConfiguration(name)
		if err != nil {
			err := fmt.Errorf("failed to delete configuration '%s' by name: %v", name, err)
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

	return resourceConfigurationRead(d, meta)
}
