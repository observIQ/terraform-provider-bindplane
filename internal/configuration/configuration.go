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

	// A list of processor names to attach to the resource
	Processors []string
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

// NewV1 takes configuration options and returns a BindPlane configuration
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

		// Build source resource with name and list
		// of processor resources
		r := model.ResourceConfiguration{
			Name: r.Name,
			ParameterizedSpec: model.ParameterizedSpec{
				Processors: processorResources,
			},
		}
		resources = append(resources, r)
	}
	return resources
}
