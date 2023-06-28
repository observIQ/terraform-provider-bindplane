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

	"github.com/observiq/bindplane-op/model"
)

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
			return fmt.Errorf("failed to set configuration labels: %v", err)
		}
		c.Metadata.Labels = l
		return nil
	}
}

// WithRawOTELConfig is a Option that configures a configuration's
// raw otel configuration.
func WithRawOTELConfig(raw string) Option {
	return func(c *model.Configuration) error {
		c.Spec.Raw = raw
		return nil
	}
}

// WithSourcesInline is a Option that configures a configuration's
// nested sources.
func WithSourcesInline(s []model.ResourceConfiguration) Option {
	return func(c *model.Configuration) error {
		if s == nil {
			return nil
		}
		c.Spec.Sources = append(c.Spec.Sources, s...)
		return nil
	}
}

// WithSourcesByName is a Option that configures a configuration's
// sources.
// func WithSourcesByName(s []string) Option {
// 	return func(c *model.Configuration) error {
// 		if s == nil {
// 			return nil
// 		}
// 		for _, s := range s {
// 			r := model.ResourceConfiguration{
// 				Name: s,
// 			}
// 			c.Spec.Sources = append(c.Spec.Sources, r)
// 		}
// 		return nil
// 	}
// }

// WithDestinationsByName is a Option that configures a configuration's
// destinations.
func WithDestinationsByName(d []string) Option {
	return func(c *model.Configuration) error {
		if d == nil {
			return nil
		}
		for _, d := range d {
			r := model.ResourceConfiguration{
				Name: d,
			}
			c.Spec.Destinations = append(c.Spec.Destinations, r)
		}
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
		kind        = "Configuration"
		contentType = "text/yaml"
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
