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
	"github.com/observiq/terraform-provider-bindplane/internal/component"
	"github.com/observiq/terraform-provider-bindplane/internal/configuration"
	"github.com/observiq/terraform-provider-bindplane/internal/maputil"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConfigurationV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceConfigurationV2Create,
		Update: resourceConfigurationV2Create, // Run create as update
		Read:   resourceConfigurationV2Read,
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
				Description: "Labels that BindPlane OP uses to determine which agents the configuration should apply to. This value is computed by Terraform and is not user configurable.",
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
						"route": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"telemetry_type": {
										Type:        schema.TypeString,
										Optional:    true,
										ForceNew:    false,
										Description: "The telemetry type to route. Valid route types include 'logs', 'metrics', or 'traces' 'logs+metrics', 'logs+traces', 'metrics+traces', 'logs+metrics+traces'.",
										ValidateFunc: func(val any, _ string) (warns []string, errs []error) {
											telemetryType := val.(string)
											if err := component.ValidateRouteType(telemetryType); err != nil {
												errs = append(errs, err)
											}
											return
										},
										Default: component.RouteTypeLogsMetricsTraces,
									},
									"components": {
										Type:        schema.TypeList,
										Required:    true,
										ForceNew:    false,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Description: "List of component names to route.",
									},
								},
							},
							Description: "Route telemetry to specific components.",
						},
					},
				},
				Description: "Source name and list of processor names to attach to the configuration. This option can be configured one or many times.",
			},
			"processor_group": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"route_id": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "The ID to use for routing to this processor group.",
						},
						"processors": {
							Type:        schema.TypeList,
							Required:    true,
							ForceNew:    false,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of processor names to attach to the processor group.",
						},
						"route": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"telemetry_type": {
										Type:        schema.TypeString,
										Optional:    true,
										ForceNew:    false,
										Description: "The telemetry type to route. Valid route types include 'logs', 'metrics', or 'traces' 'logs+metrics', 'logs+traces', 'metrics+traces', 'logs+metrics+traces'.",
										ValidateFunc: func(val any, _ string) (warns []string, errs []error) {
											telemetryType := val.(string)
											if err := component.ValidateRouteType(telemetryType); err != nil {
												errs = append(errs, err)
											}
											return
										},
										Default: component.RouteTypeLogsMetricsTraces,
									},
									"components": {
										Type:        schema.TypeList,
										Required:    true,
										ForceNew:    false,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Description: "List of component names to route.",
									},
								},
							},
							Description: "Route telemetry to specific components.",
						},
					},
				},
				Description: "Group of processors that will receive and process telemetry from one or more routes.",
			},
			"destination": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"route_id": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "The ID to use for routing to this destination.",
						},
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
				Description: "The interval at which the agent will push throughput measurements to BindPlane. Valid values include 10s, 1m, and 15m. Relaxing the interval will reduce BindPlane's measurement processing overhead at the expense of granularity. Generally, configurations with thousands of agents can justify using an interval of 1m or 15m.",
			},
			"rollout": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "Whether or not to trigger a rollout automatically when a configuration is updated. When set to true, BindPlane OP will automatically roll out the configuration change to managed agents.",
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

func resourceConfigurationV2Create(d *schema.ResourceData, meta any) error {
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
					processors = append(processors, v.(string))
				}
			}

			routes := &model.Routes{}
			if rawRoutes := sourcesRaw["route"].([]any); v != nil {
				for _, r := range rawRoutes {
					rawComponents := r.(map[string]any)["components"].([]any)
					components := []model.ComponentPath{}
					for _, c := range rawComponents {
						components = append(components, model.ComponentPath(c.(string)))
					}
					if err := component.ValidateRouteComponents(components); err != nil {
						return fmt.Errorf("validate route components: %v", err)
					}

					telemetryType := r.(map[string]any)["telemetry_type"].(string)
					switch telemetryType {
					case "logs":
						routes.Logs = append(routes.Logs, model.Route{
							Components: components,
						})
					case "metrics":
						routes.Metrics = append(routes.Metrics, model.Route{
							Components: components,
						})
					case "traces":
						routes.Traces = append(routes.Traces, model.Route{
							Components: components,
						})
					case "logs+metrics":
						routes.LogsMetrics = append(routes.LogsMetrics, model.Route{
							Components: components,
						})
					case "logs+traces":
						routes.LogsTraces = append(routes.LogsTraces, model.Route{
							Components: components,
						})
					case "metrics+traces":
						routes.MetricsTraces = append(routes.MetricsTraces, model.Route{
							Components: components,
						})
					case "logs+metrics+traces":
						routes.LogsMetricsTraces = append(routes.LogsMetricsTraces, model.Route{
							Components: components,
						})
					}
				}
			}

			sourceConf := configuration.ResourceConfig{
				Name:       sourcesRaw["name"].(string),
				Processors: processors,
				Routes:     routes,
			}
			sources = append(sources, sourceConf)
		}
	}

	processorGroups := []configuration.ResourceConfig{}
	if d.Get("processor_group") != nil {
		processorGroupsRaw := d.Get("processor_group").([]any)
		for _, v := range processorGroupsRaw {
			processorGroupRaw := v.(map[string]any)

			processors := []string{}
			if v := processorGroupRaw["processors"].([]any); v != nil {
				for _, v := range v {
					processors = append(processors, v.(string))
				}
			}

			routes := &model.Routes{}
			if rawRoutes := processorGroupRaw["route"].([]any); v != nil {
				for _, r := range rawRoutes {
					rawComponents := r.(map[string]any)["components"].([]any)
					components := []model.ComponentPath{}
					for _, c := range rawComponents {
						components = append(components, model.ComponentPath(c.(string)))
					}
					if err := component.ValidateRouteComponents(components); err != nil {
						return fmt.Errorf("validate route components: %v", err)
					}

					telemetryType := r.(map[string]any)["telemetry_type"].(string)
					switch telemetryType {
					case "logs":
						routes.Logs = append(routes.Logs, model.Route{
							Components: components,
						})
					case "metrics":
						routes.Metrics = append(routes.Metrics, model.Route{
							Components: components,
						})
					case "traces":
						routes.Traces = append(routes.Traces, model.Route{
							Components: components,
						})
					case "logs+metrics":
						routes.LogsMetrics = append(routes.LogsMetrics, model.Route{
							Components: components,
						})
					case "logs+traces":
						routes.LogsTraces = append(routes.LogsTraces, model.Route{
							Components: components,
						})
					case "metrics+traces":
						routes.MetricsTraces = append(routes.MetricsTraces, model.Route{
							Components: components,
						})
					case "logs+metrics+traces":
						routes.LogsMetricsTraces = append(routes.LogsMetricsTraces, model.Route{
							Components: components,
						})
					}
				}
			}

			processorGroupConf := configuration.ResourceConfig{
				RouteID:    processorGroupRaw["route_id"].(string),
				Processors: processors,
				Routes:     routes,
			}
			processorGroups = append(processorGroups, processorGroupConf)
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
					processors = append(processors, v.(string))
				}
			}

			destConfig := configuration.ResourceConfig{
				RouteID:    destinationRaw["route_id"].(string),
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
		configuration.WithProcessorGroups(processorGroups),
		configuration.WithDestinationsByName(destinations),
		configuration.WithExtensionsByName(extensions),
		configuration.WithRolloutOptions(rolloutOptions),
		configuration.WithMeasurementInterval(measurementInterval),
	}

	config, err := configuration.NewV2Beta(opts...)
	if err != nil {
		return fmt.Errorf("failed to create new configuration: %w", err)
	}

	resource := resource.AnyResourceFromConfigurationV1(config)
	ctx := context.Background()
	timeout := d.Timeout(schema.TimeoutCreate) - time.Minute
	if err := bindplane.ApplyWithRetry(ctx, timeout, &resource, rollout); err != nil {
		return err
	}

	return resourceConfigurationV2Read(d, meta)
}

func resourceConfigurationV2Read(d *schema.ResourceData, meta any) error {
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

		logRoutes := s.Routes.Logs
		if len(logRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range logRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "logs",
					"components":     r.Components,
				})
			}
			source["route"] = routes
		}
		metricRoutes := s.Routes.Metrics
		if len(metricRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range metricRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "metrics",
					"components":     r.Components,
				})
			}
			source["route"] = routes
		}
		traceRoutes := s.Routes.Traces
		if len(traceRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range traceRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "traces",
					"components":     r.Components,
				})
			}
			source["route"] = routes
		}
		logMetricRoutes := s.Routes.LogsMetrics
		if len(logMetricRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range logMetricRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "logs+metrics",
					"components":     r.Components,
				})
			}
			source["route"] = routes
		}
		logTraceRoutes := s.Routes.LogsTraces
		if len(logTraceRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range logTraceRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "logs+traces",
					"components":     r.Components,
				})
			}
			source["route"] = routes
		}
		metricTraceRoutes := s.Routes.MetricsTraces
		if len(metricTraceRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range metricTraceRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "metrics+traces",
					"components":     r.Components,
				})
			}
			source["route"] = routes
		}
		logMetricTraceRoutes := s.Routes.LogsMetricsTraces
		if len(logMetricTraceRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range logMetricTraceRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "logs+metrics+traces",
					"components":     r.Components,
				})
			}
			source["route"] = routes
		}

		sourceBlocks = append(sourceBlocks, source)
	}
	if err := d.Set("source", sourceBlocks); err != nil {
		return err
	}

	// Save the current state here so we can retrieve the saved
	// route ID
	stateProcessorGroupBlocks := d.Get("processor_group").([]any)

	processorGroupBlocks := []map[string]any{}
	for _, pg := range config.Spec.Processors {
		processorGroup := map[string]any{}
		processors := []string{}
		for _, p := range pg.Processors {
			processors = append(processors, strings.Split(p.Name, ":")[0])
		}
		processorGroup["processors"] = processors

		logRoutes := pg.Routes.Logs
		if len(logRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range logRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "logs",
					"components":     r.Components,
				})
			}
			processorGroup["route"] = routes
		}
		metricRoutes := pg.Routes.Metrics
		if len(metricRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range metricRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "metrics",
					"components":     r.Components,
				})
			}
			processorGroup["route"] = routes
		}
		traceRoutes := pg.Routes.Traces
		if len(traceRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range traceRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "traces",
					"components":     r.Components,
				})
			}
			processorGroup["route"] = routes
		}
		logMetricRoutes := pg.Routes.LogsMetrics
		if len(logMetricRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range logMetricRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "logs+metrics",
					"components":     r.Components,
				})
			}
			processorGroup["route"] = routes
		}
		logTraceRoutes := pg.Routes.LogsTraces
		if len(logTraceRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range logTraceRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "logs+traces",
					"components":     r.Components,
				})
			}
			processorGroup["route"] = routes
		}
		metricTraceRoutes := pg.Routes.MetricsTraces
		if len(metricTraceRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range metricTraceRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "metrics+traces",
					"components":     r.Components,
				})
			}
			processorGroup["route"] = routes
		}
		logMetricTraceRoutes := pg.Routes.LogsMetricsTraces
		if len(logMetricTraceRoutes) > 0 {
			routes := []map[string]any{}
			for _, r := range logMetricTraceRoutes {
				routes = append(routes, map[string]any{
					"telemetry_type": "logs+metrics+traces",
					"components":     r.Components,
				})
			}
			processorGroup["route"] = routes
		}

		// Retrieve the saved route IDs from state and copy them
		// to the new destination blocks before calling d.Set.
		for _, stateProcessorGroup := range stateProcessorGroupBlocks {
			stateProcessorGroup := stateProcessorGroup.(map[string]any)
			if stateProcessorGroup["name"] == processorGroup["name"] {
				processorGroup["route_id"] = stateProcessorGroup["route_id"]
				break
			}
		}

		processorGroupBlocks = append(processorGroupBlocks, processorGroup)
	}
	if err := d.Set("processor_group", processorGroupBlocks); err != nil {
		return err
	}

	// Save the current state here so we can retrieve the saved
	// route ID
	stateDestinationBlocks := d.Get("destination").([]any)

	destinationBlocks := []map[string]any{}
	for _, d := range config.Spec.Destinations {
		destination := map[string]any{}
		destination["name"] = strings.Split(d.Name, ":")[0]
		processors := []string{}
		for _, p := range d.Processors {
			processors = append(processors, strings.Split(p.Name, ":")[0])
		}
		destination["processors"] = processors

		// Retrieve the saved route IDs from state and copy them
		// to the new destination blocks before calling d.Set.
		for _, stateDestination := range stateDestinationBlocks {
			stateDestination := stateDestination.(map[string]any)
			if stateDestination["name"] == destination["name"] {
				destination["route_id"] = stateDestination["route_id"]
				break
			}
		}

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

	if err := resourceConfigurationV2RolloutOptionsRead(d, config.Spec.Rollout); err != nil {
		return err
	}

	measurementInterval := config.Spec.MeasurementInterval
	if err := d.Set("measurement_interval", measurementInterval); err != nil {
		return err
	}

	d.SetId(config.ID())
	return nil
}

// resourceConfigurationV2RolloutOptionsRead takes a configuration's rollout options
// and sets them in the Terraform state. This will trigger a terraform apply if the
// rollout options have changed outside of Terraform.
func resourceConfigurationV2RolloutOptionsRead(d *schema.ResourceData, rollout model.ResourceConfiguration) error {
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
