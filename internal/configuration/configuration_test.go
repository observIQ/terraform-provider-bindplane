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

	"github.com/observiq/bindplane-op-enterprise/model"
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
			"extensions",
			func() Option {
				r := []ResourceConfig{
					{
						Name:       "test",
						Processors: []string{},
					},
				}
				return WithExtensionsByName(r)
			}(),
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       model.KindConfiguration,
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
					Extensions: []model.ResourceConfiguration{
						{
							Name: "test",
							ParameterizedSpec: model.ParameterizedSpec{
								Processors: []model.ResourceConfiguration{},
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
		{
			"advanced-metrics",
			func() Option {
				return WithAdvancedParameters([]model.Parameter{
					{Name: "telemetryPort", Value: 8080},
					{Name: "telemetryLevel", Value: "detailed"},
				})
			}(),
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       model.KindConfiguration,
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
					Parameters: []model.Parameter{
						{Name: "telemetryPort", Value: 8080},
						{Name: "telemetryLevel", Value: "detailed"},
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

func TestNewV2Beta(t *testing.T) {
	cases := []struct {
		name   string
		input  Option
		expect *model.Configuration
	}{
		{
			"advanced-metrics",
			func() Option {
				return WithAdvancedParameters([]model.Parameter{
					{Name: "telemetryPort", Value: 8080},
					{Name: "telemetryLevel", Value: "detailed"},
				})
			}(),
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v2beta",
					Kind:       model.KindConfiguration,
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
					Parameters: []model.Parameter{
						{Name: "telemetryPort", Value: 8080},
						{Name: "telemetryLevel", Value: "detailed"},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := NewV2Beta(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expect, output)
		})
	}
}

func TestNewV2(t *testing.T) {
	cases := []struct {
		name   string
		input  Option
		expect *model.Configuration
	}{
		{
			"advanced-metrics",
			func() Option {
				return WithAdvancedParameters([]model.Parameter{
					{Name: "telemetryPort", Value: 8080},
					{Name: "telemetryLevel", Value: "detailed"},
				})
			}(),
			&model.Configuration{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v2",
					Kind:       model.KindConfiguration,
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
					Parameters: []model.Parameter{
						{Name: "telemetryPort", Value: 8080},
						{Name: "telemetryLevel", Value: "detailed"},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := NewV2(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expect, output)
		})
	}
}

// TestWithConnectorsByName_InlineTypeAndParameters verifies that an inline
// connector's Type and Parameters flow through into the resulting
// model.ResourceConfiguration's ParameterizedSpec, so the Bindplane SaaS UI
// can resolve inline routing connectors after terraform apply.
func TestWithConnectorsByName_InlineTypeAndParameters(t *testing.T) {
	params := []model.Parameter{
		{Name: "telemetry_types", Value: []interface{}{"Logs"}},
		{Name: "routes", Value: []interface{}{
			map[string]interface{}{"id": "Default"},
		}},
	}

	rc := []ResourceConfig{
		{
			RouteID:    "route-1",
			Name:       "connector-routing-1",
			Type:       "routing:3",
			Parameters: params,
		},
	}

	cfg, err := NewV2(WithConnectorsByName(rc))
	require.NoError(t, err)
	require.Len(t, cfg.Spec.Connectors, 1)

	c := cfg.Spec.Connectors[0]
	require.Equal(t, "connector-routing-1", c.Name)
	require.Equal(t, "routing:3", c.ParameterizedSpec.Type)
	require.Equal(t, params, c.ParameterizedSpec.Parameters)
}

// TestWithProcessorGroups_Parameters verifies that a processor group's
// Parameters flow through into the resulting model.ResourceConfiguration's
// ParameterizedSpec. This carries fields such as telemetry_types, which
// the SaaS UI needs to place the group under the correct telemetry section.
func TestWithProcessorGroups_Parameters(t *testing.T) {
	params := []model.Parameter{
		{Name: "telemetry_types", Value: []interface{}{"Logs"}},
	}

	rc := []ResourceConfig{
		{
			RouteID:    "pg-1",
			Parameters: params,
		},
	}

	cfg, err := NewV2(WithProcessorGroups(rc))
	require.NoError(t, err)
	require.Len(t, cfg.Spec.Processors, 1)

	pg := cfg.Spec.Processors[0]
	require.Equal(t, params, pg.ParameterizedSpec.Parameters)
}

// TestWithProcessorGroups_InlineProcessorRefs verifies that inner processors
// specified via ProcessorRefs carry their Type and Parameters through into
// the resulting model.ResourceConfiguration's inner processor list. Without
// this, inline batch and other typed processors embedded on a processor
// group are stripped on apply and the SaaS UI cannot render them.
func TestWithProcessorGroups_InlineProcessorRefs(t *testing.T) {
	refs := []ProcessorRef{
		{
			Name: "batch-inline",
			Type: "batch:3",
			Parameters: []model.Parameter{
				{Name: "send_batch_size", Value: 100},
			},
		},
	}

	rc := []ResourceConfig{
		{
			RouteID:       "pg-1",
			ProcessorRefs: refs,
		},
	}

	cfg, err := NewV2(WithProcessorGroups(rc))
	require.NoError(t, err)
	require.Len(t, cfg.Spec.Processors, 1)

	pg := cfg.Spec.Processors[0]
	require.Len(t, pg.ParameterizedSpec.Processors, 1)

	inner := pg.ParameterizedSpec.Processors[0]
	require.Equal(t, "batch-inline", inner.Name)
	require.Equal(t, "batch:3", inner.ParameterizedSpec.Type)
	require.Equal(t, refs[0].Parameters, inner.ParameterizedSpec.Parameters)
}

// TestWithProcessorGroups_MixedProcessorsAndRefs verifies that both
// name-only processors and richer processor refs can coexist on the same
// processor group. The final inner processor list should contain both
// forms in the order: Processors first, then ProcessorRefs.
func TestWithProcessorGroups_MixedProcessorsAndRefs(t *testing.T) {
	rc := []ResourceConfig{
		{
			RouteID:    "pg-1",
			Processors: []string{"library-filter"},
			ProcessorRefs: []ProcessorRef{
				{
					Name: "batch-inline",
					Type: "batch:3",
				},
			},
		},
	}

	cfg, err := NewV2(WithProcessorGroups(rc))
	require.NoError(t, err)
	require.Len(t, cfg.Spec.Processors, 1)

	pg := cfg.Spec.Processors[0]
	require.Len(t, pg.ParameterizedSpec.Processors, 2)
	require.Equal(t, "library-filter", pg.ParameterizedSpec.Processors[0].Name)
	require.Equal(t, "", pg.ParameterizedSpec.Processors[0].ParameterizedSpec.Type)
	require.Equal(t, "batch-inline", pg.ParameterizedSpec.Processors[1].Name)
	require.Equal(t, "batch:3", pg.ParameterizedSpec.Processors[1].ParameterizedSpec.Type)
}
