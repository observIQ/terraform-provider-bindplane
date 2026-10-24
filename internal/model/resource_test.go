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

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

// requireJSON requires that value is marshaled as expect, and that
// expect is unmarshaled as value
func requireJSON[T any](t *testing.T, value T, expect string) {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)
	require.JSONEq(t, expect, string(data))

	var actual T
	require.NoError(t, json.Unmarshal([]byte(expect), &actual))
	require.Equal(t, value, actual)
}

func TestResourceMeta(t *testing.T) {
	r := ResourceMeta{
		Metadata: Metadata{
			ID:     "01JAZ3NDEKTSV4RRFFQ69G5FAV",
			Name:   "my-source",
			Labels: LabelsFromValidatedMap(map[string]string{"env": "test"}),
		},
	}
	require.Equal(t, "01JAZ3NDEKTSV4RRFFQ69G5FAV", r.ID())
	require.Equal(t, "my-source", r.Name())
	require.Equal(t, map[string]string{"env": "test"}, r.GetLabels().AsMap())

	// Resources embed ResourceMeta
	s := Source{ResourceInstance{ResourceMeta: r}}
	require.Equal(t, r.ID(), s.ID())
	require.Equal(t, r.Name(), s.Name())
}

func TestAnyResource_JSON(t *testing.T) {
	r := AnyResource{
		ResourceMeta: ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       KindSource,
			Metadata: Metadata{
				Name:   "my-source",
				Labels: LabelsFromValidatedMap(map[string]string{"env": "test"}),
			},
		},
		Spec: map[string]any{
			"type": "otlp",
			"parameters": []any{
				map[string]any{"name": "grpc_port", "value": "4317"},
			},
		},
	}

	requireJSON(t, r, `{
		"apiVersion": "bindplane.observiq.com/v1",
		"kind": "Source",
		"metadata": {"id": "", "name": "my-source", "labels": {"env": "test"}},
		"spec": {"type": "otlp", "parameters": [{"name": "grpc_port", "value": "4317"}]}
	}`)
}

func TestAnyResourceStatus_JSON(t *testing.T) {
	cases := []struct {
		name   string
		status UpdateStatus
		reason string
	}{
		{"unchanged", StatusUnchanged, ""},
		{"configured", StatusConfigured, ""},
		{"created", StatusCreated, ""},
		{"invalid", "invalid", "missing required parameter"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := AnyResourceStatus{
				Resource: AnyResource{
					ResourceMeta: ResourceMeta{
						APIVersion: "bindplane.observiq.com/v1",
						Kind:       KindConfiguration,
						Metadata: Metadata{
							ID:     "01JAZ3NDEKTSV4RRFFQ69G5FAV",
							Name:   "my-config",
							Labels: LabelsFromValidatedMap(map[string]string{}),
						},
					},
				},
				Status: tc.status,
				Reason: tc.reason,
			}

			requireJSON(t, s, fmt.Sprintf(`{
				"resource": {
					"apiVersion": "bindplane.observiq.com/v1",
					"kind": "Configuration",
					"metadata": {"id": "01JAZ3NDEKTSV4RRFFQ69G5FAV", "name": "my-config", "labels": {}},
					"spec": null
				},
				"status": %q,
				"reason": %q
			}`, tc.status, tc.reason))
		})
	}
}

func TestResourceInstance_JSON(t *testing.T) {
	instance := func(kind Kind) ResourceInstance {
		return ResourceInstance{
			ResourceMeta: ResourceMeta{
				APIVersion: "bindplane.observiq.com/v1",
				Kind:       kind,
				Metadata: Metadata{
					ID:     "01JAZ3NDEKTSV4RRFFQ69G5FAV",
					Name:   "my-resource",
					Labels: LabelsFromValidatedMap(map[string]string{"env": "test"}),
				},
			},
			Spec: ParameterizedSpec{
				Type: "my-type:3",
				Parameters: Parameters{
					{Name: "endpoint", Value: "localhost:4317"},
					{Name: "api_key", Value: "(sensitive)", Sensitive: true},
					{Name: "password", Value: "my-credential", Credential: true},
				},
				Processors: []ResourceConfiguration{{Name: "my-processor:1"}},
			},
		}
	}

	expect := func(kind Kind) string {
		return fmt.Sprintf(`{
			"apiVersion": "bindplane.observiq.com/v1",
			"kind": %q,
			"metadata": {"id": "01JAZ3NDEKTSV4RRFFQ69G5FAV", "name": "my-resource", "labels": {"env": "test"}},
			"spec": {
				"type": "my-type:3",
				"parameters": [
					{"name": "endpoint", "value": "localhost:4317"},
					{"name": "api_key", "value": "(sensitive)", "sensitive": true},
					{"name": "password", "value": "my-credential", "credential": true}
				],
				"processors": [{"name": "my-processor:1"}]
			}
		}`, kind)
	}

	t.Run("source", func(t *testing.T) {
		requireJSON(t, Source{instance(KindSource)}, expect(KindSource))
	})
	t.Run("processor", func(t *testing.T) {
		requireJSON(t, Processor{instance(KindProcessor)}, expect(KindProcessor))
	})
	t.Run("destination", func(t *testing.T) {
		requireJSON(t, Destination{instance(KindDestination)}, expect(KindDestination))
	})
	t.Run("extension", func(t *testing.T) {
		requireJSON(t, Extension{instance(KindExtension)}, expect(KindExtension))
	})
	t.Run("connector", func(t *testing.T) {
		requireJSON(t, Connector{instance(KindConnector)}, expect(KindConnector))
	})
}

func TestConfiguration_JSON(t *testing.T) {
	c := Configuration{
		ResourceMeta: ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       KindConfiguration,
			Metadata: Metadata{
				ID:     "01JAZ3NDEKTSV4RRFFQ69G5FAV",
				Name:   "my-config",
				Labels: LabelsFromValidatedMap(map[string]string{"platform": "linux"}),
			},
		},
		Spec: ConfigurationSpec{
			ContentType:         "text/yaml",
			MeasurementInterval: "1m",
			Parameters:          Parameters{{Name: "telemetryLevel", Value: "detailed"}},
			Sources: []ResourceConfiguration{{
				ID:   "source-1",
				Name: "my-source:1",
				Routes: &Routes{
					Logs: []Route{{ID: "route-1", Components: []ComponentPath{"destinations/destination-1"}}},
				},
			}},
			Processors:   []ResourceConfiguration{{ID: "processor-1", Name: "my-processor:1"}},
			Connectors:   []ResourceConfiguration{{ID: "connector-1", Name: "my-connector:1"}},
			Destinations: []ResourceConfiguration{{ID: "destination-1", Name: "my-destination:1"}},
			Extensions:   []ResourceConfiguration{{Name: "my-extension:1"}},
			Selector:     AgentSelector{MatchLabels: MatchLabels{"configuration": "my-config"}},
			Rollout: ResourceConfiguration{
				ParameterizedSpec: ParameterizedSpec{
					Type:       "progressive:1",
					Parameters: Parameters{{Name: "stages", Value: []any{"stage-1"}}},
				},
			},
		},
	}

	requireJSON(t, c, `{
		"apiVersion": "bindplane.observiq.com/v1",
		"kind": "Configuration",
		"metadata": {"id": "01JAZ3NDEKTSV4RRFFQ69G5FAV", "name": "my-config", "labels": {"platform": "linux"}},
		"spec": {
			"contentType": "text/yaml",
			"measurementInterval": "1m",
			"parameters": [{"name": "telemetryLevel", "value": "detailed"}],
			"sources": [{
				"id": "source-1",
				"name": "my-source:1",
				"routes": {"logs": [{"id": "route-1", "components": ["destinations/destination-1"]}]}
			}],
			"processors": [{"id": "processor-1", "name": "my-processor:1"}],
			"connectors": [{"id": "connector-1", "name": "my-connector:1"}],
			"destinations": [{"id": "destination-1", "name": "my-destination:1"}],
			"extensions": [{"name": "my-extension:1"}],
			"selector": {"matchLabels": {"configuration": "my-config"}},
			"rollout": {"type": "progressive:1", "parameters": [{"name": "stages", "value": ["stage-1"]}]}
		}
	}`)

	t.Run("empty spec", func(t *testing.T) {
		c := Configuration{
			ResourceMeta: ResourceMeta{
				APIVersion: "bindplane.observiq.com/v1",
				Kind:       KindConfiguration,
				Metadata:   Metadata{Name: "my-config", Labels: LabelsFromValidatedMap(map[string]string{})},
			},
		}

		// Omitted fields are not sent to Bindplane
		requireJSON(t, c, `{
			"apiVersion": "bindplane.observiq.com/v1",
			"kind": "Configuration",
			"metadata": {"id": "", "name": "my-config", "labels": {}},
			"spec": {"contentType": "", "measurementInterval": "", "selector": {"matchLabels": null}, "rollout": {}}
		}`)
	})
}

func TestFleet_JSON(t *testing.T) {
	f := Fleet{
		ResourceMeta: ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       KindFleet,
			Metadata: Metadata{
				ID:          "my-fleet",
				Name:        "my-fleet",
				DisplayName: "My Fleet",
				Labels:      LabelsFromValidatedMap(map[string]string{"agent-type": "observiq-otel-collector", "platform": "linux"}),
			},
		},
		FleetSpec: FleetSpec{
			Configuration: "my-config",
			Selector:      AgentSelector{MatchLabels: MatchLabels{"fleet": "my-fleet"}},
		},
	}

	requireJSON(t, f, `{
		"apiVersion": "bindplane.observiq.com/v1",
		"kind": "Fleet",
		"metadata": {
			"id": "my-fleet",
			"name": "my-fleet",
			"displayName": "My Fleet",
			"labels": {"agent-type": "observiq-otel-collector", "platform": "linux"}
		},
		"spec": {"configuration": "my-config", "selector": {"matchLabels": {"fleet": "my-fleet"}}}
	}`)
}

func TestFleetSpec_mapstructure(t *testing.T) {
	// The provider converts the fleet spec to the spec of an AnyResource,
	// which must have the same JSON representation as the fleet spec
	spec := FleetSpec{
		Configuration: "my-config",
		Selector:      AgentSelector{MatchLabels: MatchLabels{"fleet": "my-fleet"}},
	}

	var specMap map[string]any
	require.NoError(t, mapstructure.Decode(spec, &specMap))

	data, err := json.Marshal(specMap)
	require.NoError(t, err)
	require.JSONEq(t, `{"configuration": "my-config", "selector": {"matchLabels": {"fleet": "my-fleet"}}}`, string(data))
}
