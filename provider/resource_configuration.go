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

	"github.com/observiq/terraform-provider-bindplane/internal/client"
	"github.com/observiq/terraform-provider-bindplane/internal/configuration"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceConfigurationCreate,
		Update: resourceConfigurationCreate, // Run create as update
		Read:   resourceConfigurationRead,
		Delete: genericConfigurationDelete,
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
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
						"processors": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: false,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"destination": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
						"processors": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: false,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
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
		return err
	}
	labels["platform"] = d.Get("platform").(string)

	// Match labels should always be configuration=<name>
	matchLabels := map[string]string{
		"configuration": name,
	}

	// List of sources and their processors
	sources := []configuration.ResourceConfig{}
	if d.Get("source") != nil {
		sourcesRaw := d.Get("source").([]any)
		for _, v := range sourcesRaw {
			sourcesRaw := v.(map[string]any)

			processors := []string{}
			if v := sourcesRaw["processors"].(*schema.Set); v != nil {
				for _, processorName := range v.List() {
					processors = append(processors, processorName.(string))
				}
			}

			sourceConf := configuration.ResourceConfig{
				Name:       sourcesRaw["name"].(string),
				Processors: processors,
			}
			sources = append(sources, sourceConf)
		}
	}

	// List of destinations and their processors
	destinations := []configuration.ResourceConfig{}
	if d.Get("destination") != nil {
		destinationsRaw := d.Get("destination").([]any)
		for _, v := range destinationsRaw {
			destinationRaw := v.(map[string]any)

			processors := []string{}
			if v := destinationRaw["processors"].(*schema.Set); v != nil {
				for _, processorName := range v.List() {
					processors = append(processors, processorName.(string))
				}
			}

			destConfig := configuration.ResourceConfig{
				Name:       destinationRaw["name"].(string),
				Processors: processors,
			}
			destinations = append(destinations, destConfig)
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

	resource := resource.AnyResourceFromConfigurationV1(config)
	bindplane := meta.(*client.BindPlane)
	ctx := context.TODO()
	timeout := d.Timeout(schema.TimeoutCreate) - time.Minute
	if err := bindplane.ApplyWithRetry(ctx, timeout, &resource, rollout); err != nil {
		return err
	}

	return resourceConfigurationRead(d, meta)
}

func resourceConfigurationRead(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	config, err := bindplane.Configuration(d.Get("name").(string))
	if err != nil {
		return err
	}

	if config == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("name", config.Name()); err != nil {
		return err
	}

	labels := config.Metadata.Labels.AsMap()
	platform, ok := labels["platform"]
	if ok {
		if err := d.Set("platform", platform); err != nil {
			return err
		}
		// Remove the platform label from the labels map
		// because Terraform's state does not expect it.
		delete(labels, "platform")
	}

	// Save the labels map to state, which has the 'platform' label removed.
	if err := d.Set("labels", labels); err != nil {
		return err
	}

	matchLabels := make(map[string]string)
	for k, v := range config.Spec.Selector.MatchLabels {
		matchLabels[k] = v
	}
	if err := d.Set("match_labels", matchLabels); err != nil {
		return err
	}

	d.SetId(config.ID())
	return nil
}
