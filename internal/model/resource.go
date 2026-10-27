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

// Package model provides the Bindplane resources used by the provider. It
// contains only the types and fields the provider uses, with the same JSON
// representation as the Bindplane API.
package model

// Kind is the kind of a Bindplane resource
type Kind string

// Kinds of Bindplane resources
const (
	KindConfiguration Kind = "Configuration"
	KindAgent         Kind = "Agent"
	KindSource        Kind = "Source"
	KindProcessor     Kind = "Processor"
	KindConnector     Kind = "Connector"
	KindDestination   Kind = "Destination"
	KindExtension     Kind = "Extension"
	KindFleet         Kind = "Fleet"
	KindGateway       Kind = "Gateway"
)

// ResourceMeta contains the API version, kind, and metadata shared by all
// Bindplane resources
type ResourceMeta struct {
	APIVersion string   `json:"apiVersion"`
	Kind       Kind     `json:"kind"`
	Metadata   Metadata `json:"metadata"`
}

// ID returns the ID of the resource
func (r *ResourceMeta) ID() string {
	return r.Metadata.ID
}

// Name returns the name of the resource
func (r *ResourceMeta) Name() string {
	return r.Metadata.Name
}

// GetLabels returns the labels of the resource
func (r *ResourceMeta) GetLabels() Labels {
	return r.Metadata.Labels
}

// Metadata is the metadata of a resource
type Metadata struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	Labels      Labels `json:"labels"`
}

// AnyResource is a resource of any kind, with an unstructured spec
type AnyResource struct {
	ResourceMeta
	Spec map[string]any `json:"spec"`
}

// UpdateStatus is the result of applying a resource
type UpdateStatus string

// Statuses of applied resources which are not errors
const (
	StatusUnchanged  UpdateStatus = "unchanged"
	StatusConfigured UpdateStatus = "configured"
	StatusCreated    UpdateStatus = "created"
)

// AnyResourceStatus is the result of applying a resource
type AnyResourceStatus struct {
	Resource AnyResource  `json:"resource"`
	Status   UpdateStatus `json:"status"`
	Reason   string       `json:"reason"`
}

// ResourceInstance is a source, processor, destination, extension,
// or connector
type ResourceInstance struct {
	ResourceMeta
	Spec ParameterizedSpec `json:"spec"`
}

// Source is a Bindplane source
type Source struct {
	ResourceInstance
}

// Processor is a Bindplane processor
type Processor struct {
	ResourceInstance
}

// Destination is a Bindplane destination
type Destination struct {
	ResourceInstance
}

// Extension is a Bindplane extension
type Extension struct {
	ResourceInstance
}

// Connector is a Bindplane connector
type Connector struct {
	ResourceInstance
}

// Configuration is a Bindplane configuration
type Configuration struct {
	ResourceMeta
	Spec ConfigurationSpec `json:"spec"`
}

// ConfigurationSpec is the spec of a configuration
type ConfigurationSpec struct {
	ContentType         string                  `json:"contentType"`
	MeasurementInterval string                  `json:"measurementInterval"`
	Parameters          Parameters              `json:"parameters,omitempty"`
	Sources             []ResourceConfiguration `json:"sources,omitempty"`
	Processors          []ResourceConfiguration `json:"processors,omitempty"`
	Connectors          []ResourceConfiguration `json:"connectors,omitempty"`
	Destinations        []ResourceConfiguration `json:"destinations,omitempty"`
	Extensions          []ResourceConfiguration `json:"extensions,omitempty"`
	Selector            AgentSelector           `json:"selector"`
	Rollout             ResourceConfiguration   `json:"rollout,omitempty"`
}

// Fleet is a Bindplane fleet
type Fleet struct {
	ResourceMeta
	FleetSpec `json:"spec"`
}

// FleetSpec is the spec of a fleet. The mapstructure tags are used when
// the provider converts the spec to a map.
type FleetSpec struct {
	// Configuration is the name of the configuration assigned to the fleet
	Configuration string `json:"configuration" mapstructure:"configuration"`

	// Selector matches agents to the fleet
	Selector AgentSelector `json:"selector" mapstructure:"selector"`
}
