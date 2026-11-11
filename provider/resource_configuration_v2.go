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

	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/observiq/terraform-provider-bindplane/internal/component"
	"github.com/observiq/terraform-provider-bindplane/internal/configuration"
	"github.com/observiq/terraform-provider-bindplane/internal/maputil"
	"github.com/observiq/terraform-provider-bindplane/internal/model"
	"github.com/observiq/terraform-provider-bindplane/internal/parameter"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"
	v2 "github.com/observiq/terraform-provider-bindplane/provider/resource/configuration/v2"

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
			"description": configurationSchemaDescription(),
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
							Description: "List of processor names to attach to the source. Conflicts with processor.",
						},
						"processor": processorBlockSchema(),
						"route":     v2.RouteSchema,
					},
				},
				Description: "Source name and list of processor names to attach to the configuration. This option can be configured one or many times.",
			},
			"connector": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"route_id": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "The ID to use for routing to this connector.",
						},
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    false,
							Description: "Name of the connector to attach. Conflicts with type.",
						},
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    false,
							Description: "The type of an inline connector, such as 'routing'. Conflicts with name.",
						},
						"parameters_json": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         false,
							Description:      "A JSON array of parameters used to configure an inline connector. Requires type.",
							DiffSuppressFunc: suppressEquivalentJSONDiffs,
						},
						"route": v2.RouteSchema,
					},
				},
				Description: "Connector to be attached to the configuration. This option can be configured one or many times.",
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
						"parameters_json": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         false,
							Description:      "A JSON array of parameters used to configure the processor group, such as telemetry_types.",
							DiffSuppressFunc: suppressEquivalentJSONDiffs,
						},
						"processors": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of processor names to attach to the processor group. Conflicts with processor.",
						},
						"processor": processorBlockSchema(),
						"route":     v2.RouteSchema,
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
							Description: "List of processor names to attach to the destination. Conflicts with processor.",
						},
						"processor": processorBlockSchema(),
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
				Description: "The interval at which the agent will push throughput measurements to Bindplane. Valid values include 10s, 1m, and 15m. Relaxing the interval will reduce Bindplane's measurement processing overhead at the expense of granularity. Generally, configurations with thousands of agents can justify using an interval of 1m or 15m.",
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
			"advanced": advancedSchema,
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

			processors, err := processorsFromConfig(sourcesRaw["processors"].([]any), sourcesRaw["processor"].([]any))
			if err != nil {
				return fmt.Errorf("source '%s': %w", sourcesRaw["name"], err)
			}

			routes := &model.Routes{}
			if rawRoutes := sourcesRaw["route"].(*schema.Set).List(); v != nil {
				r, err := component.ParseRoutes(rawRoutes)
				if err != nil {
					return fmt.Errorf("parse routes: %w", err)
				}
				routes = r
			}

			sourceConf := configuration.ResourceConfig{
				Name:       sourcesRaw["name"].(string),
				Processors: processors,
				Routes:     routes,
			}
			sources = append(sources, sourceConf)
		}
	}

	connectors := []configuration.ResourceConfig{}
	if d.Get("connector") != nil {
		connectorsRaw := d.Get("connector").([]any)
		for _, v := range connectorsRaw {
			connectorRaw := v.(map[string]any)

			routeID := connectorRaw["route_id"].(string)
			connectorName := connectorRaw["name"].(string)
			connectorType := connectorRaw["type"].(string)

			parameters, err := parseInlineComponent(connectorName, connectorType, connectorRaw["parameters_json"].(string))
			if err != nil {
				return fmt.Errorf("connector '%s': %w", routeID, err)
			}

			routes := &model.Routes{}
			if rawRoutes := connectorRaw["route"].(*schema.Set).List(); v != nil {
				r, err := component.ParseRoutes(rawRoutes)
				if err != nil {
					return fmt.Errorf("parse routes: %w", err)
				}
				routes = r
			}

			connectorConf := configuration.ResourceConfig{
				RouteID:    routeID,
				Name:       connectorName,
				Type:       connectorType,
				Parameters: parameters,
				Routes:     routes,
			}
			connectors = append(connectors, connectorConf)
		}
	}

	processorGroups := []configuration.ResourceConfig{}
	if d.Get("processor_group") != nil {
		processorGroupsRaw := d.Get("processor_group").([]any)
		for _, v := range processorGroupsRaw {
			processorGroupRaw := v.(map[string]any)

			routeID := processorGroupRaw["route_id"].(string)

			var parameters []model.Parameter
			if s := processorGroupRaw["parameters_json"].(string); s != "" {
				params, err := parameter.StringToParameter(s)
				if err != nil {
					return fmt.Errorf("processor group '%s': %w", routeID, err)
				}
				parameters = params
			}

			processors, err := processorsFromConfig(processorGroupRaw["processors"].([]any), processorGroupRaw["processor"].([]any))
			if err != nil {
				return fmt.Errorf("processor group '%s': %w", routeID, err)
			}

			routes := &model.Routes{}
			if rawRoutes := processorGroupRaw["route"].(*schema.Set).List(); v != nil {
				r, err := component.ParseRoutes(rawRoutes)
				if err != nil {
					return fmt.Errorf("parse routes: %w", err)
				}
				routes = r
			}

			processorGroupConf := configuration.ResourceConfig{
				RouteID:    routeID,
				Parameters: parameters,
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

			processors, err := processorsFromConfig(destinationRaw["processors"].([]any), destinationRaw["processor"].([]any))
			if err != nil {
				return fmt.Errorf("destination '%s': %w", destinationRaw["name"], err)
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

	// Extract advanced parameters
	advancedParameters, err := extractAdvancedParameters(d)
	if err != nil {
		return fmt.Errorf("failed to extract advanced parameters: %w", err)
	}

	opts := []configuration.Option{
		configuration.WithName(name),
		configuration.WithDescription(d.Get("description").(string)),
		configuration.WithLabels(labels),
		configuration.WithMatchLabels(matchLabels),
		configuration.WithSourcesByName(sources),
		configuration.WithConnectorsByName(connectors),
		configuration.WithProcessorGroups(processorGroups),
		configuration.WithDestinationsByName(destinations),
		configuration.WithExtensionsByName(extensions),
		configuration.WithRolloutOptions(rolloutOptions),
		configuration.WithMeasurementInterval(measurementInterval),
		configuration.WithAdvancedParameters(advancedParameters),
	}

	config, err := configuration.NewV2(opts...)
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

	// Always set the description, including the empty string, so that
	// a description removed outside of Terraform is detected as drift.
	if err := d.Set("description", config.Metadata.Description); err != nil {
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

	// Save the current state here so we can reconcile the
	// processors of each source
	stateSourceBlocks := d.Get("source").([]any)

	sourceBlocks := []map[string]any{}
	for _, s := range config.Spec.Sources {
		source := map[string]any{}
		source["name"] = strings.Split(s.Name, ":")[0]

		stateSource := stateBlock(stateSourceBlocks, "name", source["name"])
		processors, processorBlocks, err := processorsToState(stateSource, s.Processors)
		if err != nil {
			return fmt.Errorf("source '%s': %w", source["name"], err)
		}
		source["processors"] = processors
		source["processor"] = processorBlocks

		stateRoutes, err := component.RoutesToState(s.Routes)
		if err != nil {
			return fmt.Errorf("routes to state: %w", err)
		}
		source["route"] = stateRoutes

		sourceBlocks = append(sourceBlocks, source)
	}
	if err := d.Set("source", sourceBlocks); err != nil {
		return err
	}

	// Save the current state here so we can reconcile the
	// parameters of each connector
	stateConnectorBlocks := d.Get("connector").([]any)

	connectorBlocks := []map[string]any{}
	for _, c := range config.Spec.Connectors {
		connector := map[string]any{}
		connector["route_id"] = c.ID
		connector["name"] = strings.Split(c.Name, ":")[0]
		connector["type"] = strings.Split(c.Type, ":")[0]

		stateParams, _ := stateBlock(stateConnectorBlocks, "route_id", c.ID)["parameters_json"].(string)
		params, err := reconcileParametersJSON(stateParams, c.Parameters)
		if err != nil {
			return fmt.Errorf("connector '%s': %w", c.ID, err)
		}
		connector["parameters_json"] = params

		stateRoutes, err := component.RoutesToState(c.Routes)
		if err != nil {
			return fmt.Errorf("routes to state: %w", err)
		}
		connector["route"] = stateRoutes

		connectorBlocks = append(connectorBlocks, connector)
	}
	if err := d.Set("connector", connectorBlocks); err != nil {
		return err
	}

	// Save the current state here so we can retrieve the saved
	// route ID
	stateProcessorGroupBlocks := d.Get("processor_group").([]any)

	processorGroupBlocks := []map[string]any{}
	for _, pg := range config.Spec.Processors {
		processorGroup := map[string]any{}

		stateProcessorGroup := stateBlock(stateProcessorGroupBlocks, "route_id", pg.ID)

		stateParams, _ := stateProcessorGroup["parameters_json"].(string)
		params, err := reconcileParametersJSON(stateParams, pg.Parameters)
		if err != nil {
			return fmt.Errorf("processor group '%s': %w", pg.ID, err)
		}
		processorGroup["parameters_json"] = params

		processors, processorBlocks, err := processorsToState(stateProcessorGroup, pg.Processors)
		if err != nil {
			return fmt.Errorf("processor group '%s': %w", pg.ID, err)
		}
		processorGroup["processors"] = processors
		processorGroup["processor"] = processorBlocks

		stateRoutes, err := component.RoutesToState(pg.Routes)
		if err != nil {
			return fmt.Errorf("routes to state: %w", err)
		}
		processorGroup["route"] = stateRoutes

		// Retrieve the saved route IDs from state and copy them
		// to the new processor blocks before calling d.Set.
		if stateProcessorGroup != nil {
			processorGroup["route_id"] = stateProcessorGroup["route_id"]
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

		stateDestination := stateBlock(stateDestinationBlocks, "name", destination["name"])
		processors, processorBlocks, err := processorsToState(stateDestination, d.Processors)
		if err != nil {
			return fmt.Errorf("destination '%s': %w", destination["name"], err)
		}
		destination["processors"] = processors
		destination["processor"] = processorBlocks

		// Retrieve the saved route IDs from state and copy them
		// to the new destination blocks before calling d.Set.
		if stateDestination != nil {
			destination["route_id"] = stateDestination["route_id"]
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

	if err := resourceConfigurationRolloutOptionsRead(d, config.Spec.Rollout); err != nil {
		return err
	}

	measurementInterval := config.Spec.MeasurementInterval
	if err := d.Set("measurement_interval", measurementInterval); err != nil {
		return err
	}

	if err := setAdvancedMetricsInState(d, config); err != nil {
		return err
	}

	d.SetId(config.ID())
	return nil
}

// processorBlockSchema returns the schema for processors that are
// attached to a source, processor group, or destination. Processors
// are referenced by name or defined inline with a type and parameters.
func processorBlockSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: false,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    false,
					Description: "Name of the processor to attach. Conflicts with type.",
				},
				"type": {
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    false,
					Description: "The type of an inline processor, such as 'batch'. Conflicts with name.",
				},
				"parameters_json": {
					Type:             schema.TypeString,
					Optional:         true,
					ForceNew:         false,
					Description:      "A JSON array of parameters used to configure an inline processor. Requires type.",
					DiffSuppressFunc: suppressEquivalentJSONDiffs,
				},
				"display_name": {
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    false,
					Description: "The name of the processor displayed in the Bindplane UI.",
				},
				"id": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					ForceNew:    false,
					Description: "The ID of an inline processor. When unset, the ID generated by Bindplane is saved and reused by later applies.",
				},
			},
		},
		Description: "Processor to attach, referenced by name or defined inline with a type and parameters. Conflicts with processors. This option can be configured one or many times.",
	}
}

// parseInlineComponent validates a component that is either referenced by
// name or defined inline by type, and returns the parameters of an inline
// component.
func parseInlineComponent(name, componentType, parametersJSON string) ([]model.Parameter, error) {
	switch {
	case name != "" && componentType != "":
		return nil, errors.New("name and type cannot both be set")
	case name == "" && componentType == "":
		return nil, errors.New("one of name or type must be set")
	case parametersJSON == "":
		return nil, nil
	case componentType == "":
		return nil, errors.New("parameters_json requires type")
	}
	return parameter.StringToParameter(parametersJSON)
}

// processorsFromConfig returns the processors configured by a block's
// processors list, which references processors by name, or by its
// processor blocks, which reference processors by name or define them
// inline. Both cannot be set.
func processorsFromConfig(names, blocks []any) ([]model.ResourceConfiguration, error) {
	if len(names) > 0 && len(blocks) > 0 {
		return nil, errors.New("processors and processor cannot both be set")
	}

	processors := []model.ResourceConfiguration{}
	for _, name := range names {
		processors = append(processors, model.ResourceConfiguration{Name: name.(string)})
	}

	for i, b := range blocks {
		// An empty processor block is nil. The type
		// assertions below return zero values for it.
		block, _ := b.(map[string]any)
		name, _ := block["name"].(string)
		processorType, _ := block["type"].(string)
		parametersJSON, _ := block["parameters_json"].(string)
		displayName, _ := block["display_name"].(string)

		parameters, err := parseInlineComponent(name, processorType, parametersJSON)
		if err != nil {
			return nil, fmt.Errorf("processor %d: %w", i, err)
		}

		processor := model.ResourceConfiguration{
			Name:        name,
			DisplayName: displayName,
			ParameterizedSpec: model.ParameterizedSpec{
				Type:       processorType,
				Parameters: parameters,
			},
		}

		// Bindplane derives the ID of a processor referenced by name
		// from its name. Only inline processors keep the ID from state,
		// which is empty when Bindplane has not generated one yet.
		if processorType != "" {
			processor.ID, _ = block["id"].(string)
		}

		processors = append(processors, processor)
	}

	return processors, nil
}

// processorsToState returns the processors of a source, processor group,
// or destination as a list of names and a list of processor blocks, one
// of which is empty. Processor blocks are returned when the block saved
// to state used them, or when a processor is defined inline, because an
// inline processor cannot be represented by name. The parameters of each
// processor are reconciled with the processor block at the same position
// in state.
func processorsToState(stateBlock map[string]any, processors []model.ResourceConfiguration) ([]string, []map[string]any, error) {
	stateProcessors, _ := stateBlock["processor"].([]any)

	useBlocks := len(stateProcessors) > 0
	for _, p := range processors {
		if p.Name == "" {
			useBlocks = true
		}
	}

	if !useBlocks {
		names := []string{}
		for _, p := range processors {
			names = append(names, strings.Split(p.Name, ":")[0])
		}
		return names, nil, nil
	}

	blocks := []map[string]any{}
	for i, p := range processors {
		stateParams := ""
		if i < len(stateProcessors) {
			stateProcessor, _ := stateProcessors[i].(map[string]any)
			stateParams, _ = stateProcessor["parameters_json"].(string)
		}

		params, err := reconcileParametersJSON(stateParams, p.Parameters)
		if err != nil {
			return nil, nil, fmt.Errorf("processor %d: %w", i, err)
		}

		blocks = append(blocks, map[string]any{
			"id":              p.ID,
			"name":            strings.Split(p.Name, ":")[0],
			"type":            strings.Split(p.Type, ":")[0],
			"parameters_json": params,
			"display_name":    p.DisplayName,
		})
	}

	return nil, blocks, nil
}

// stateBlock returns the block saved to state whose key is equal
// to value, or nil if there is no such block.
func stateBlock(blocks []any, key string, value any) map[string]any {
	for _, b := range blocks {
		if block, ok := b.(map[string]any); ok && block[key] == value {
			return block
		}
	}
	return nil
}
