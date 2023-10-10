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

// Package resource provides functions for defining bindplane
// generic resources.

package parameter

import (
	"strings"
	"testing"

	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/stretchr/testify/require"
)

// import (
// 	"testing"

// 	"github.com/observiq/bindplane-op-enterprise/model"
// 	"github.com/stretchr/testify/require"
// )

func TestStringToParameter(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		expect    []model.Parameter
		expectErr string
	}{
		{
			"no-params",
			"",
			nil,
			"",
		},
		{
			"valid",
			`[
				{
					"name":"project",
					"value":"my-gcp-project"
				}
			]`,
			[]model.Parameter{
				{
					Name:  "project",
					Value: "my-gcp-project",
				},
			},
			"",
		},
		{
			"invalid-json",
			`[
				{
					"name",
					"value":"my-gcp-project"
				}
			]`,
			[]model.Parameter{
				{
					Name:  "project",
					Value: "my-gcp-project",
				},
			},
			"failed to unmarshal parameters",
		},
		{
			"multi-valid",
			`[
				{
				  "name": "collection_interval",
				  "value": 20
				},
				{
				  "name": "enable_process",
				  "value": false
				},
				{
				  "name": "metric_filtering",
				  "value": [
					"system.disk.operation_time"
				  ]
				},
				{
					"name":"string",
					"value":"value"
				},
				{
					"name":"map",
					"value": {
						"bool":true,
						"num":401,
						"x":"y"
					}
				},
				{
					"name":"project",
					"value":"my-gcp-project"
				}
			]`,
			[]model.Parameter{
				{
					Name:  "collection_interval",
					Value: float64(20),
				},
				{
					Name:  "enable_process",
					Value: false,
				},
				{
					Name:  "metric_filtering",
					Value: []any{"system.disk.operation_time"},
				},
				{
					Name:  "string",
					Value: "value",
				},
				{
					Name: "map",
					Value: map[string]any{
						"bool": true,
						"num":  float64(401),
						"x":    "y",
					},
				},
				{
					Name:  "project",
					Value: "my-gcp-project",
				},
			},
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := StringToParameter(tc.input)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.Equal(t, tc.expect, output)
		})
	}
}

func TestParametersToSring(t *testing.T) {
	cases := []struct {
		name      string
		input     []model.Parameter
		expect    string
		expectErr string
	}{
		{
			"no-params",
			nil,
			"",
			"",
		},
		{
			"valid",
			[]model.Parameter{
				{
					Name:  "project",
					Value: "my-gcp-project",
				},
			},
			`[
				{
					"name":"project",
					"value":"my-gcp-project"
				}
			]`,
			"",
		},
		{
			"multi-valid",
			[]model.Parameter{
				{
					Name:  "collection_interval",
					Value: float64(20),
				},
				{
					Name:  "enable_process",
					Value: false,
				},
				{
					Name:  "metric_filtering",
					Value: []any{"system.disk.operation_time"},
				},
				{
					Name:  "string",
					Value: "value",
				},
				{
					Name: "map",
					Value: map[string]any{
						"bool": true,
						"num":  float64(401),
						"x":    "y",
					},
				},
				{
					Name:  "project",
					Value: "my-gcp-project",
				},
			},
			`[
				{
				  "name": "collection_interval",
				  "value": 20
				},
				{
				  "name": "enable_process",
				  "value": false
				},
				{
				  "name": "metric_filtering",
				  "value": [
					"system.disk.operation_time"
				  ]
				},
				{
					"name":"string",
					"value":"value"
				},
				{
					"name":"map",
					"value": {
						"bool":true,
						"num":401,
						"x":"y"
					}
				},
				{
					"name":"project",
					"value":"my-gcp-project"
				}
			]`,
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expect := tc.expect
			expect = strings.Replace(expect, " ", "", -1)
			expect = strings.Replace(expect, "\t", "", -1)
			expect = strings.Replace(expect, "\n", "", -1)

			output, err := ParametersToString(tc.input)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.Equal(t, expect, output)
		})
	}
}
