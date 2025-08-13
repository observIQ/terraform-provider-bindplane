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

	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/stretchr/testify/require"
)

func TestAnyResourceV1(t *testing.T) {
	cases := []struct {
		name            string
		id              string
		rName           string
		rType           string
		rkind           model.Kind
		rParameters     []model.Parameter
		rProcessors     []model.ResourceConfiguration
		extraSpecFields map[string]any
		expectErr       string
	}{
		{
			"source",
			"tf-source",
			"my-host",
			"host",
			model.KindSource,
			[]model.Parameter{
				{
					Name:  "collection_interval",
					Value: 60,
				},
				{
					Name:  "tags",
					Value: []string{"test"},
				},
			},
			nil,
			nil,
			"",
		},
		{
			"destination",
			"tf-destination",
			"my-destination",
			"googlecloud",
			model.KindDestination,
			[]model.Parameter{
				{
					Name:  "project",
					Value: "my-project",
				},
			},
			nil,
			nil,
			"",
		},
		{
			"processor",
			"tf-processor",
			"my-filter",
			"filter",
			model.KindProcessor,
			nil,
			nil,
			nil,
			"",
		},
		{
			"extension",
			"tf-extension",
			"my-extension",
			"pprof",
			model.KindExtension,
			nil,
			nil,
			nil,
			"",
		},
		{
			"invalid-kind",
			"tf-resource",
			"my-resource",
			"resource",
			model.KindAgent,
			nil,
			nil,
			nil,
			"unknown bindplane resource kind: Agent",
		},
		{
			"valid-processors",
			"tf-bundle",
			"my-bundle",
			"bundle",
			model.KindProcessor,
			nil,
			[]model.ResourceConfiguration{
				{
					Name: "filter-a",
				},
				{
					Name: "filter-b",
				},
			},
			nil,
			"",
		},
		{
			"processor-with-recommendation",
			"tf-processor-rec",
			"my-filter-rec",
			"filter",
			model.KindProcessor,
			nil,
			nil,
			map[string]any{
				"recommendation": "test-recommendation",
			},
			"",
		},
		{
			"processor-with-nil-extra-fields",
			"tf-processor-nil",
			"my-filter-nil",
			"filter",
			model.KindProcessor,
			nil,
			nil,
			nil,
			"",
		},
		{
			"processor-with-empty-recommendation",
			"tf-processor-empty",
			"my-filter-empty",
			"filter",
			model.KindProcessor,
			nil,
			nil,
			nil, // Should be nil when recommendation is empty string
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := AnyResourceV1(tc.id, tc.rName, tc.rType, tc.rkind, tc.rParameters, tc.rProcessors, tc.extraSpecFields)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestAnyResourceV1WithExtraSpecFields(t *testing.T) {
	// Test that extra spec fields are properly included
	resource, err := AnyResourceV1(
		"test-id",
		"test-processor",
		"filter",
		model.KindProcessor,
		nil,
		nil,
		map[string]any{
			"recommendation": "test-recommendation",
		},
	)
	require.NoError(t, err)
	require.NotNil(t, resource)

	// Verify the recommendation field is in the spec
	recommendation, exists := resource.Spec["recommendation"]
	require.True(t, exists, "recommendation field should exist in spec")
	require.Equal(t, "test-recommendation", recommendation)

	// Verify other standard fields are still present
	require.Equal(t, "filter", resource.Spec["type"])
	require.Equal(t, []model.Parameter(nil), resource.Spec["parameters"])
	require.Equal(t, []map[string]string{}, resource.Spec["processors"])
}

func TestAnyResourceV1WithNilExtraSpecFields(t *testing.T) {
	// Test that nil extra spec fields don't cause issues
	resource, err := AnyResourceV1(
		"test-id",
		"test-processor",
		"filter",
		model.KindProcessor,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, resource)

	// Verify standard fields are present
	require.Equal(t, "filter", resource.Spec["type"])
	require.Equal(t, []model.Parameter(nil), resource.Spec["parameters"])
	require.Equal(t, []map[string]string{}, resource.Spec["processors"])

	// Verify no extra fields are added
	_, exists := resource.Spec["recommendation"]
	require.False(t, exists, "recommendation field should not exist when not provided")
}

func TestAnyResourceV1WithEmptyExtraSpecFields(t *testing.T) {
	// Test that empty string values in extra spec fields are filtered out
	resource, err := AnyResourceV1(
		"test-id",
		"test-processor",
		"filter",
		model.KindProcessor,
		nil,
		nil,
		map[string]any{
			"recommendation": "",
			"other_field":    "non-empty-value",
		},
	)
	require.NoError(t, err)
	require.NotNil(t, resource)

	// Verify empty string values are not included
	_, exists := resource.Spec["recommendation"]
	require.False(t, exists, "empty string values should not be included in spec")

	// Verify non-empty values are included
	otherField, exists := resource.Spec["other_field"]
	require.True(t, exists, "non-empty values should be included in spec")
	require.Equal(t, "non-empty-value", otherField)

	// Verify standard fields are still present
	require.Equal(t, "filter", resource.Spec["type"])
	require.Equal(t, []model.Parameter(nil), resource.Spec["parameters"])
	require.Equal(t, []map[string]string{}, resource.Spec["processors"])
}

func TestAnyResourceFromConfigurationV1(t *testing.T) {
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
					Kind:       model.KindConfiguration,
				},
				Spec: model.ConfigurationSpec{
					ContentType: "text/yaml",
				},
			},
			model.AnyResource{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       model.KindConfiguration,
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
					Kind:       model.KindConfiguration,
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
					Extensions: []model.ResourceConfiguration{
						{
							Name: "pprof",
						},
					},
				},
			},
			model.AnyResource{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       model.KindConfiguration,
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
					"extensions": []model.ResourceConfiguration{
						{
							Name: "pprof",
						},
					},
				},
			},
		},
		{
			"advanced-parameters",
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
			model.AnyResource{
				ResourceMeta: model.ResourceMeta{
					APIVersion: "bindplane.observiq.com/v1",
					Kind:       model.KindConfiguration,
				},
				Spec: map[string]any{
					"contentType": "text/yaml",
					"selector": model.AgentSelector{
						MatchLabels: nil,
					},
					"parameters": model.Parameters{
						{Name: "telemetryPort", Value: 8080},
						{Name: "telemetryLevel", Value: "detailed"},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := AnyResourceFromConfigurationV1(tc.input)
			require.Equal(t, tc.expect, output)
		})
	}
}
