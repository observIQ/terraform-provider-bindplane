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

package model

// Parameter is the name and value of a resource parameter
type Parameter struct {
	Name  string `json:"name"`
	Value any    `json:"value"`

	// Sensitive is true if the value is masked by Bindplane
	Sensitive bool `json:"sensitive,omitempty"`

	// Credential is true if the value is a credential reference
	Credential bool `json:"credential,omitempty"`
}

// Parameters is a list of parameters
type Parameters []Parameter

// ParameterizedSpec is the spec of a resource with a type and parameters
type ParameterizedSpec struct {
	Type       string                  `json:"type,omitempty"`
	Parameters Parameters              `json:"parameters,omitempty"`
	Processors []ResourceConfiguration `json:"processors,omitempty"`
}

// ResourceConfiguration is a reference to a resource by name, or an
// embedded resource, within another resource
type ResourceConfiguration struct {
	// ID identifies the resource when routing
	ID string `json:"id,omitempty"`

	// Name is the name of the referenced resource
	Name string `json:"name,omitempty"`

	// DisplayName is the name displayed in the Bindplane UI
	DisplayName string `json:"displayName,omitempty"`

	ParameterizedSpec

	// Routes connect resources in a configuration
	Routes *Routes `json:"routes,omitempty"`
}

// Routes connect a resource to other resources, by telemetry type
type Routes struct {
	Logs    []Route `json:"logs,omitempty"`
	Metrics []Route `json:"metrics,omitempty"`
	Traces  []Route `json:"traces,omitempty"`
}

// Route connects a resource to a list of components
type Route struct {
	ID         string          `json:"id,omitempty"`
	Components []ComponentPath `json:"components,omitempty"`
}

// ComponentPath identifies a component in a configuration,
// e.g. destinations/my-destination
type ComponentPath string

// AgentSelector selects agents by label. The mapstructure tag is used
// when the provider converts a fleet spec to a map.
type AgentSelector struct {
	MatchLabels `json:"matchLabels" mapstructure:"matchLabels"`
}

// MatchLabels are the labels an agent must have to be selected
type MatchLabels map[string]string
