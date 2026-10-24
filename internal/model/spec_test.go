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
	"testing"
)

func TestParameter_JSON(t *testing.T) {
	cases := []struct {
		name   string
		param  Parameter
		expect string
	}{
		{"string", Parameter{Name: "endpoint", Value: "localhost:4317"}, `{"name": "endpoint", "value": "localhost:4317"}`},
		{"bool", Parameter{Name: "enable_tls", Value: true}, `{"name": "enable_tls", "value": true}`},
		{"list", Parameter{Name: "files", Value: []any{"/var/log/a.log", "/var/log/b.log"}}, `{"name": "files", "value": ["/var/log/a.log", "/var/log/b.log"]}`},
		{"null", Parameter{Name: "endpoint"}, `{"name": "endpoint", "value": null}`},
		{"sensitive", Parameter{Name: "api_key", Value: "(sensitive)", Sensitive: true}, `{"name": "api_key", "value": "(sensitive)", "sensitive": true}`},
		{"credential", Parameter{Name: "password", Value: "my-credential", Credential: true}, `{"name": "password", "value": "my-credential", "credential": true}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			requireJSON(t, tc.param, tc.expect)
		})
	}
}

func TestResourceConfiguration_JSON(t *testing.T) {
	cases := []struct {
		name   string
		config ResourceConfiguration
		expect string
	}{
		{
			name:   "empty",
			config: ResourceConfiguration{},
			expect: `{}`,
		},
		{
			name:   "reference by name",
			config: ResourceConfiguration{ID: "source-1", Name: "my-source:1"},
			expect: `{"id": "source-1", "name": "my-source:1"}`,
		},
		{
			// The spec of an embedded resource is inline with its ID
			name: "embedded resource",
			config: ResourceConfiguration{
				ID: "processor-1",
				ParameterizedSpec: ParameterizedSpec{
					Type:       "batch:1",
					Parameters: Parameters{{Name: "send_batch_size", Value: "8192"}},
					Processors: []ResourceConfiguration{{Name: "my-processor:1"}},
				},
			},
			expect: `{
				"id": "processor-1",
				"type": "batch:1",
				"parameters": [{"name": "send_batch_size", "value": "8192"}],
				"processors": [{"name": "my-processor:1"}]
			}`,
		},
		{
			name: "routes",
			config: ResourceConfiguration{
				ID:   "source-1",
				Name: "my-source:1",
				Routes: &Routes{
					Logs:    []Route{{ID: "logs", Components: []ComponentPath{"destinations/destination-1"}}},
					Metrics: []Route{{ID: "metrics", Components: []ComponentPath{"processors/processor-1", "connectors/connector-1"}}},
					Traces:  []Route{{Components: []ComponentPath{"destinations/destination-2"}}},
				},
			},
			expect: `{
				"id": "source-1",
				"name": "my-source:1",
				"routes": {
					"logs": [{"id": "logs", "components": ["destinations/destination-1"]}],
					"metrics": [{"id": "metrics", "components": ["processors/processor-1", "connectors/connector-1"]}],
					"traces": [{"components": ["destinations/destination-2"]}]
				}
			}`,
		},
		{
			name:   "empty routes",
			config: ResourceConfiguration{Name: "my-source:1", Routes: &Routes{}},
			expect: `{"name": "my-source:1", "routes": {}}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			requireJSON(t, tc.config, tc.expect)
		})
	}
}

func TestAgentSelector_JSON(t *testing.T) {
	cases := []struct {
		name     string
		selector AgentSelector
		expect   string
	}{
		{"nil", AgentSelector{}, `{"matchLabels": null}`},
		{"labels", AgentSelector{MatchLabels: MatchLabels{"configuration": "my-config"}}, `{"matchLabels": {"configuration": "my-config"}}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			requireJSON(t, tc.selector, tc.expect)
		})
	}
}
