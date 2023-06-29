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

package resource

import (
	"testing"

	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

func TestAnyResourceFromConfiguration(t *testing.T) {
	cases := []struct {
		name   string
		input  *model.Configuration
		expect model.AnyResource
	}{
		{
			"valid-no-resources",
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       "Configuration",
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
				},
			},
			model.AnyResource{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       "Configuration",
				},
				Spec: map[string]any{
					"contentType": "text/yaml",
					"selector": model.AgentSelector{
						MatchLabels: nil,
					},
				},
			},
		},
		{
			"valid-resources",
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       "Configuration",
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
					Sources: []model.ResourceConfiguration{
						{
							Name: "source-a",
						},
						{
							ParameterizedSpec: model.ParameterizedSpec{
								Type: "host",
								Parameters: []model.Parameter{
									{
										Name:  "collection_interval",
										Value: "60",
									},
								},
							},
						},
					},
					Destinations: []model.ResourceConfiguration{
						{
							Name: "logging",
						},
						{
							ParameterizedSpec: model.ParameterizedSpec{
								Type: "custom",
								Parameters: []model.Parameter{
									{
										Name:  "configuration",
										Value: "logging:",
									},
								},
							},
						},
					},
				},
			},
			model.AnyResource{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       "Configuration",
				},
				Spec: map[string]any{
					"contentType": "text/yaml",
					"selector": model.AgentSelector{
						MatchLabels: nil,
					},
					"sources": []model.ResourceConfiguration{
						{
							Name: "source-a",
						},
						{
							ParameterizedSpec: model.ParameterizedSpec{
								Type: "host",
								Parameters: []model.Parameter{
									{
										Name:  "collection_interval",
										Value: "60",
									},
								},
							},
						},
					},
					"destinations": []model.ResourceConfiguration{
						{
							Name: "logging",
						},
						{
							ParameterizedSpec: model.ParameterizedSpec{
								Type: "custom",
								Parameters: []model.Parameter{
									{
										Name:  "configuration",
										Value: "logging:",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := AnyResourceFromConfiguration(tc.input)
			require.Equal(t, tc.expect, output)
		})
	}
}

func TestAnyResourceFromRawConfiguration(t *testing.T) {
	cases := []struct {
		name   string
		input  *model.Configuration
		expect model.AnyResource
	}{
		{
			"empty",
			&model.Configuration{},
			model.AnyResource{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "",
					Kind:       "",
					Metadata:   model.Metadata{},
				},
				Spec: map[string]any{
					"contentType": "",
					"raw":         "",
					"selector":    model.AgentSelector{},
				},
			},
		},
		{
			"set",
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "v0",
					Kind:       "TestConf",
					Metadata: model.Metadata{
						Name: "test",
					},
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/test",
					Raw:         "OTEL",
					Selector: model.AgentSelector{
						MatchLabels: model.MatchLabels{
							"key": "value",
						},
					},
				},
			},
			model.AnyResource{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "v0",
					Kind:       "TestConf",
					Metadata: model.Metadata{
						Name: "test",
					},
				},
				Spec: map[string]any{
					"contentType": "text/test",
					"raw":         "OTEL",
					"selector": model.AgentSelector{
						MatchLabels: model.MatchLabels{
							"key": "value",
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := AnyResourceFromRawConfiguration(tc.input)
			require.Equal(t, tc.expect, output)
		})
	}
}
