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

// Package configuration provides functions for defining bindplane
// configuration resources.
package configuration

import (
	"fmt"

	"github.com/observiq/bindplane-op-enterprise/model"
)

// ResourceConfig represents the configuration of a destination
// or source that will be attached to a configuration.
type ResourceConfig struct {
	// The name of the resource (destination / source) that should be
	// attached to the configuration
	Name string

	// A list of processor names to attach to the resource. Used for
	// library-referenced processors whose type and parameters live on
	// a separate bindplane_processor resource.
	Processors []string

	// ProcessorRefs is a richer representation of attached processors.
	// In addition to a name, each ref can carry Type and Parameters for
	// inline processors whose metadata lives on the configuration spec
	// rather than on a separate bindplane_processor resource. Both
	// Processors and ProcessorRefs are merged into the final attached
	// processor list on apply.
	ProcessorRefs []ProcessorRef

	// RouteID is the ID to use when routing to this resource
	RouteID string

	// Routes to attach to the resource
	Routes *model.Routes

	// Type is the component type carried inline on the configuration
	// spec, for example "routing:3" on an inline routing connector.
	// Empty for library-referenced components, whose type is carried
	// on the referenced resource rather than the configuration spec.
	Type string

	// Parameters are component parameters carried inline on the
	// configuration spec. Used by inline connectors for routing
	// conditions and by processor groups for fields such as
	// telemetry_types.
	Parameters []model.Parameter
}

// ProcessorRef represents a processor attached to a source, processor
// group, or destination. Unlike the bare Processors name list it can
// carry Type and Parameters for inline processors whose metadata lives
// on the configuration spec rather than on a separate
// bindplane_processor resource.
type ProcessorRef struct {
	// Name is the name of the processor to attach.
	Name string

	// Type is the component type for an inline processor, for example
	// "batch:3". Empty for library-referenced processors whose type is
	// carried on the referenced bindplane_processor resource.
	Type string

	// Parameters are component parameters carried inline on the
	// configuration spec.
	Parameters []model.Parameter
}

// Option is a function that configures a
// bindplane model.Configuration
type Option func(*model.Configuration) error

// WithName is a Option that configures a configuration's
// name.
func WithName(name string) Option {
	return func(c *model.Configuration) error {
		c.Metadata.Name = name
		return nil
	}
}

// WithLabels is a Option that configures a configuration's
// labels.
func WithLabels(labels map[string]string) Option {
	return func(c *model.Configuration) error {
		l, err := model.LabelsFromMap(labels)
		if err != nil {
			return fmt.Errorf("failed to set configuration labels: %w", err)
		}
		c.Metadata.Labels = l
		return nil
	}
}

// WithSourcesByName is a Option that configures a configuration's
// sources.
func WithSourcesByName(s []ResourceConfig) Option {
	return func(c *model.Configuration) error {
		c.Spec.Sources = append(c.Spec.Sources, withResourcesByName(s)...)
		return nil
	}
}

// WithConnectorsByName is a Option that configures a configuration's
// connectors.
func WithConnectorsByName(connector []ResourceConfig) Option {
	return func(c *model.Configuration) error {
		c.Spec.Connectors = append(c.Spec.Connectors, withResourcesByName(connector)...)
		return nil
	}
}

// WithProcessorGroups is a Option that configures a configuration's
// processor groups.
func WithProcessorGroups(p []ResourceConfig) Option {
	return func(c *model.Configuration) error {
		c.Spec.Processors = append(c.Spec.Processors, withResourcesByName(p)...)
		return nil
	}
}

// WithDestinationsByName is a Option that configures a configuration's
// destinations.
func WithDestinationsByName(d []ResourceConfig) Option {
	return func(c *model.Configuration) error {
		c.Spec.Destinations = append(c.Spec.Destinations, withResourcesByName(d)...)
		return nil
	}
}

// WithExtensionsByName is a Option that configures a configuration's
// extensions.
func WithExtensionsByName(d []ResourceConfig) Option {
	return func(c *model.Configuration) error {
		c.Spec.Extensions = append(c.Spec.Extensions, withResourcesByName(d)...)
		return nil
	}
}

// WithMatchLabels is a Option that configures a configuration's
// agent match labels.
func WithMatchLabels(match map[string]string) Option {
	return func(c *model.Configuration) error {
		c.Spec.Selector.MatchLabels = match
		return nil
	}
}

// WithRolloutOptions takes a model.ResourceConfiguration and returns
// an Option that configures a configuration's rollout options. It is safe
// to pass model.ResourceConfiguration's zero value to this function.
func WithRolloutOptions(rolloutOptions model.ResourceConfiguration) Option {
	return func(c *model.Configuration) error {
		c.Spec.Rollout = rolloutOptions
		return nil
	}
}

// WithMeasurementInterval is a Option that configures a configuration's
// measurement interval.
func WithMeasurementInterval(interval string) Option {
	return func(c *model.Configuration) error {
		// Validation is not performed here because Terraform
		// schema validation will already ensure the value is
		// a valid duration acceptable by Bindplane.
		c.Spec.MeasurementInterval = interval
		return nil
	}
}

// WithAdvancedParameters is an Option that configures a configuration's
// advanced parameters.
func WithAdvancedParameters(parameters []model.Parameter) Option {
	return func(c *model.Configuration) error {
		c.Spec.Parameters = append(c.Spec.Parameters, parameters...)
		return nil
	}
}

// NewV1 takes configuration options and returns a Bindplane configuration
func NewV1(options ...Option) (*model.Configuration, error) {
	const (
		version     = "bindplane.observiq.com/v1"
		kind        = model.KindConfiguration
		contentType = "text/yaml" // TODO(jsirianni): Is this required and does it make sense?
	)

	c := &model.Configuration{
		ResourceMeta: model.ResourceMeta{
			APIVersion: version,
			Kind:       kind,
		},
		Spec: model.ConfigurationSpec{
			ContentType: contentType,
		},
	}

	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// NewV2 wraps NewV2Beta and returns a BindPlane configuration
// with API version bindplane.observiq.com/v2.
func NewV2(options ...Option) (*model.Configuration, error) {
	c, err := NewV2Beta(options...)
	if err != nil {
		return nil, err
	}
	c.ResourceMeta.APIVersion = "bindplane.observiq.com/v2"
	return c, nil
}

// NewV2Beta takes a configuration options and returns a BindPlane configuration
// with API version bindplane.observiq.com/v2beta
func NewV2Beta(options ...Option) (*model.Configuration, error) {
	const (
		version     = "bindplane.observiq.com/v2beta"
		kind        = model.KindConfiguration
		contentType = "text/yaml" // TODO(jsirianni): Is this required and does it make sense?
	)

	c := &model.Configuration{
		ResourceMeta: model.ResourceMeta{
			APIVersion: version,
			Kind:       kind,
		},
		Spec: model.ConfigurationSpec{
			ContentType: contentType,
		},
	}

	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// WithResourcesByName takes a list of resource configurations
// and returns a list of bindplane model.ResourceConfigurations.
func withResourcesByName(r []ResourceConfig) []model.ResourceConfiguration {
	resources := []model.ResourceConfiguration{}

	for _, r := range r {
		// build list of processor resource objects by name
		processorResources := []model.ResourceConfiguration{}
		for _, name := range r.Processors {
			processor := model.ResourceConfiguration{
				Name: name,
			}
			processorResources = append(processorResources, processor)
		}
		// richer processor refs with optional inline Type / Parameters
		for _, ref := range r.ProcessorRefs {
			processorResources = append(processorResources, model.ResourceConfiguration{
				Name: ref.Name,
				ParameterizedSpec: model.ParameterizedSpec{
					Type:       ref.Type,
					Parameters: ref.Parameters,
				},
			})
		}

		routeID := r.RouteID

		// Build source resource with name and list
		// of processor resources
		r := model.ResourceConfiguration{
			Name: r.Name,
			ParameterizedSpec: model.ParameterizedSpec{
				Type:       r.Type,
				Parameters: r.Parameters,
				Processors: processorResources,
			},
			Routes: r.Routes,
		}
		if routeID != "" {
			r.ID = routeID
		}

		resources = append(resources, r)
	}
	return resources
}
