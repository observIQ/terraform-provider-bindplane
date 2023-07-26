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

package configuration

import (
	"testing"

	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

func TestNewV1(t *testing.T) {
	cases := []struct {
		name   string
		input  Option
		expect *model.Configuration
	}{
		{
			"no-options",
			nil,
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       model.KindConfiguration,
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
				},
			},
		},
		{
			"name",
			WithName("observiq"),
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       model.KindConfiguration,
					Metadata: model.Metadata{
						Name: "observiq",
					},
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
				},
			},
		},
		{
			"raw-config",
			WithRawOTELConfig("some raw config"),
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       model.KindConfiguration,
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
					Raw:         "some raw config",
				},
			},
		},
		{
			"sources",
			func() Option {
				r := []ResourceConfig{
					{
						Name: "test",
						Processors: []string{
							"count",
							"batch",
						},
					},
				}
				return WithSourcesByName(r)
			}(),
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       model.KindConfiguration,
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
					Sources: []model.ResourceConfiguration{
						{
							Name: "test",
							ParameterizedSpec: model.ParameterizedSpec{
								Processors: []model.ResourceConfiguration{
									{
										Name: "count",
									},
									{
										Name: "batch",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"destinations",
			func() Option {
				r := []ResourceConfig{
					{
						Name: "test",
						Processors: []string{
							"count",
							"batch",
						},
					},
				}
				return WithDestinationsByName(r)
			}(),
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       model.KindConfiguration,
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
					Destinations: []model.ResourceConfiguration{
						{
							Name: "test",
							ParameterizedSpec: model.ParameterizedSpec{
								Processors: []model.ResourceConfiguration{
									{
										Name: "count",
									},
									{
										Name: "batch",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"labels",
			func() Option {
				labels := map[string]string{
					"key":  "value",
					"test": "withLabels",
				}
				return WithLabels(labels)
			}(),
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       model.KindConfiguration,
					Metadata: model.Metadata{
						Labels: func() model.Labels {
							in := map[string]string{
								"key":  "value",
								"test": "withLabels",
							}
							l, _ := model.LabelsFromMap(in)
							return l
						}(),
					},
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
				},
			},
		},
		{
			"match-labels",
			func() Option {
				match := map[string]string{
					"matchkey":  "value",
					"matchtest": "withLabels",
				}
				return WithMatchLabels(match)
			}(),
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       model.KindConfiguration,
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
					Selector: model.AgentSelector{
						MatchLabels: map[string]string{
							"matchkey":  "value",
							"matchtest": "withLabels",
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := NewV1(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expect, output)
		})
	}
}
