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
	"strings"
	"time"

	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/observiq/terraform-provider-bindplane/internal/configuration"
	"github.com/observiq/terraform-provider-bindplane/internal/maputil"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the configuration.",
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
				Description: "The platform the configuration is for.",
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
				ValidateFunc: func(val any, _ string) (warns []string, errs []error) {
					labels := val.(map[string]any)
					_, ok := labels["platform"]
					if ok {
						errs = append(errs, errors.New("label 'platform' will be overwritten by the configured platform"))
					}
					return
				},
				Description: "Key value pairs which will be added to the configuration as labels.",
			},
			"match_labels": {
				Type:        schema.TypeMap,
				Computed:    true,
				ForceNew:    false,
				Description: "Labels that Bindplane uses to determine which agents the configuration should apply to. This value is computed by Terraform and is not user configurable.",
			},
			"source": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "Name of the source to attach.",
						},
						"processors": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of processor names to attach to the source.",
						},
					},
				},
				Description: "Source name and list of processor names to attach to the configuration. This option can be configured one or many times.",
			},
			"destination": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "Name of the destination to attach.",
						},
						"processors": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of processor names to attach to the destination.",
						},
					},
				},
				Description: "Destination name and list of processor names to attach to the configuration. This option can be configured one or many times.",
			},
			"extensions": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    false,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of extensions names to attach to the configuration.",
			},
			"measurement_interval": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				ValidateFunc: func(val any, _ string) (warns []string, errs []error) {
					interval := val.(string)
					if interval != "10s" && interval != "1m" && interval != "15m" {
						errs = append(errs, errors.New("measurement_interval must be one of 10s, 1m, or 15m"))
					}
					return
				},
				Description: "The interval at which the agent will push throughput measurements to Bindplane. Valid values include 10s, 1m, and 15m. Relaxing the interval will reduce BindPlane's measurement processing overhead at the expense of granularity. Generally, configurations with thousands of agents can justify using an interval of 1m or 15m.",
			},
			"rollout": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "Whether or not to trigger a rollout automatically when a configuration is updated. When set to true, Bindplane will automatically roll out the configuration change to managed agents.",
			},
			"rollout_options": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(val any, _ string) (warns []string, errs []error) {
								t := val.(string)
								if t != "standard" && t != "progressive" {
									errs = append(errs, fmt.Errorf("invalid rollout type: %s", t))
								}
								return
							},
							ForceNew:    false,
							Description: "The type of rollout to perform. Valid values are 'standard' and 'progressive'.",
						},
						"parameters": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Name of the parameter.",
									},
									"value": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"labels": {
													Type:        schema.TypeMap,
													Required:    true,
													Description: "Labels for the parameter.",
												},
												"name": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Name of the stage.",
												},
											},
										},
										Description: "Value of the parameter, which is a list of stages.",
									},
								},
							},
							Description: "List of parameters for the rollout options.",
						},
					},
				},
				Description: "Options for configuring the rollout behavior of the configuration.",
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
	bindplane := meta.(*client.BindPlane)

	name := d.Get("name").(string)
	rollout := d.Get("rollout").(bool)

	// If id is unset, it means Terraform has not previously created
	// this resource. Check to ensure a resource with this name does
	// not already exist.
	if d.Id() == "" {
		c, err := bindplane.Configuration(name)
		if err != nil {
			return err
		}
		if c != nil {
			return fmt.Errorf("configuration with name '%s' already exists with id '%s'", name, c.ID())
		}
	}

	labels, err := maputil.StringMapFromTFMap(d.Get("labels").(map[string]any))
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
			if v := sourcesRaw["processors"].([]any); v != nil {
				for _, v := range v {
					processorName := v.(string)
					proc, err := bindplane.Processor(processorName)
					if err != nil {
						return fmt.Errorf("failed to check processor type: %w", err)
					}
					if proc != nil && model.TrimVersion(proc.Spec.Type) == "processor_bundle" {
						// If this is a bundle, check its subprocessors for any processor bundles
						if err := checkForProcessorBundlesInSubprocessors(bindplane, processorName); err != nil {
							return err
						}
					}
					processors = append(processors, processorName)
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
			if v := destinationRaw["processors"].([]any); v != nil {
				for _, v := range v {
					processorName := v.(string)
					proc, err := bindplane.Processor(processorName)
					if err != nil {
						return fmt.Errorf("failed to check processor type: %w", err)
					}

					if proc != nil && proc.Spec.Type == "processor_bundle" {
						// If this is a bundle, check its subprocessors for any processor bundles
						if err := checkForProcessorBundlesInSubprocessors(bindplane, processorName); err != nil {
							return err
						}
					}
					processors = append(processors, processorName)
				}
			}

			destConfig := configuration.ResourceConfig{
				Name:       destinationRaw["name"].(string),
				Processors: processors,
			}
			destinations = append(destinations, destConfig)
		}
	}

	rolloutOptions, err := readRolloutOptions(d)
	if err != nil {
		return fmt.Errorf("read rollout_options: %w", err)
	}

	// List of extensions represented as a list of configuration.ResourceConfig's
	// with only the name field set.
	extensions := []configuration.ResourceConfig{}
	if e := d.Get("extensions").([]any); e != nil {
		for _, extension := range e {
			extensionConfig := configuration.ResourceConfig{
				Name: extension.(string),
			}
			extensions = append(extensions, extensionConfig)
		}
	}

	measurementInterval := d.Get("measurement_interval").(string)

	opts := []configuration.Option{
		configuration.WithName(name),
		configuration.WithLabels(labels),
		configuration.WithMatchLabels(matchLabels),
		configuration.WithSourcesByName(sources),
		configuration.WithDestinationsByName(destinations),
		configuration.WithExtensionsByName(extensions),
		configuration.WithRolloutOptions(rolloutOptions),
		configuration.WithMeasurementInterval(measurementInterval),
	}

	config, err := configuration.NewV1(opts...)
	if err != nil {
		return fmt.Errorf("failed to create new configuration: %w", err)
	}

	resource := resource.AnyResourceFromConfigurationV1(config)
	ctx := context.Background()
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

	// If the state ID is set but differs from the ID returned by,
	// bindplane, mark the resource to be re-created by unsetting
	// the ID. This will cause Terraform to attempt to create the resource
	// instead of updating it. The creation step will fail because
	// the resource already exists. This behavior is desirable, it will
	// prevent Terraform from modifying resources created by other means.
	if id := d.Id(); id != "" {
		if config.ID() != d.Id() {
			d.SetId("")
			return nil
		}
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

	sourceBlocks := []map[string]any{}
	for _, s := range config.Spec.Sources {
		source := map[string]any{}
		source["name"] = strings.Split(s.Name, ":")[0]
		processors := []string{}
		for _, p := range s.Processors {
			processors = append(processors, strings.Split(p.Name, ":")[0])
		}
		source["processors"] = processors
		sourceBlocks = append(sourceBlocks, source)
	}
	if err := d.Set("source", sourceBlocks); err != nil {
		return err
	}

	destinationBlocks := []map[string]any{}
	for _, d := range config.Spec.Destinations {
		destination := map[string]any{}
		destination["name"] = strings.Split(d.Name, ":")[0]
		processors := []string{}
		for _, p := range d.Processors {
			processors = append(processors, strings.Split(p.Name, ":")[0])
		}
		destination["processors"] = processors
		destinationBlocks = append(destinationBlocks, destination)
	}
	if err := d.Set("destination", destinationBlocks); err != nil {
		return err
	}

	extensions := []string{}
	for _, e := range config.Spec.Extensions {
		extensions = append(extensions, strings.Split(e.Name, ":")[0])
	}
	if err := d.Set("extensions", extensions); err != nil {
		return err
	}

	if err := resourceConfigurationRolloutOptionsRead(d, config.Spec.Rollout); err != nil {
		return err
	}

	measurementInterval := config.Spec.MeasurementInterval
	if err := d.Set("measurement_interval", measurementInterval); err != nil {
		return err
	}

	d.SetId(config.ID())
	return nil
}

// resourceConfigurationRolloutOptionsRead takes a configuration's rollout options
// and sets them in the Terraform state. This will trigger a terraform apply if the
// rollout options have changed outside of Terraform.
func resourceConfigurationRolloutOptionsRead(d *schema.ResourceData, rollout model.ResourceConfiguration) error {
	if len(rollout.Parameters) == 0 {
		return nil
	}

	rolloutOptions := make(map[string]interface{})

	rolloutOptions["type"] = rollout.Type

	parameters := make([]interface{}, len(rollout.Parameters))
	for i, param := range rollout.Parameters {
		parameters[i] = map[string]interface{}{
			"name":  param.Name,
			"value": param.Value,
		}
	}

	rolloutOptions["parameters"] = parameters

	if err := d.Set("rollout_options", []interface{}{rolloutOptions}); err != nil {
		return fmt.Errorf("error setting rollout options: %s", err)
	}

	return nil
}

func checkForProcessorBundlesInSubprocessors(bindplane *client.BindPlane, processorName string) error {
	proc, err := bindplane.Processor(processorName)
	if err != nil {
		return fmt.Errorf("failed to check processor type: %w", err)
	}

	if proc == nil {
		return nil
	}

	// If this is a processor bundle, check its subprocessors
	if model.TrimVersion(proc.Spec.Type) == "processor_bundle" {
		bundle, err := bindplane.Processor(processorName)
		if err != nil {
			return fmt.Errorf("failed to get processor bundle: %w", err)
		}
		// Check each subprocessor
		for _, subProc := range bundle.Spec.Processors {
			subProcessor, err := bindplane.Processor(subProc.Name)
			if err != nil {
				return fmt.Errorf("failed to get subprocessor information: %w", err)
			}

			if subProcessor != nil && model.TrimVersion(subProcessor.Spec.Type) == "processor_bundle" {
				return fmt.Errorf("nested processor bundles are not supported: processor bundle '%s' contains another processor bundle '%s' as a subprocessor", processorName, subProc.Name)
			}
		}
	}

	return nil
}
