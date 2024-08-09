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
		name        string
		rName       string
		rType       string
		rkind       model.Kind
		rParameters []model.Parameter
		expectErr   string
	}{
		{
			"source",
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
			"",
		},
		{
			"destination",
			"my-destination",
			"googlecloud",
			model.KindDestination,
			[]model.Parameter{
				{
					Name:  "project",
					Value: "my-project",
				},
			},
			"",
		},
		{
			"processor",
			"my-filter",
			"filter",
			model.KindProcessor,
			nil,
			"",
		},
		{
			"extension",
			"my-extension",
			"pprof",
			model.KindExtension,
			nil,
			"",
		},
		{
			"invalid-kind",
			"my-resource",
			"resource",
			model.KindAgent,
			nil,
			"unknown bindplane resource kind: Agent",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := AnyResourceV1(tc.rName, tc.rType, tc.rkind, tc.rParameters, "", "")
			if tc.expectErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
		})
	}
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := AnyResourceFromConfigurationV1(tc.input)
			require.Equal(t, tc.expect, output)
		})
	}
}
