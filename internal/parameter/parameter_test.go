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
	"testing"

	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

func TestStringToParameter(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		expect    []model.Parameter
		expectErr string
	}{
		{
			"valid",
			`{"project":"my-gcp-project"}`,
			[]model.Parameter{
				{
					Name:  "project",
					Value: "my-gcp-project",
				},
			},
			"",
		},
		{
			"multi-valid",
			`{"project":"my-gcp-project","a":"b","int":12,"map":{"x":"y","num":401,"bool":true}}`,
			[]model.Parameter{
				{
					Name:  "a",
					Value: "b",
				},
				{
					Name:  "int",
					Value: float64(12),
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
		{
			"invalid",
			`{"project"}`,
			nil,
			"failed to convert string parameters to map[string]any",
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
			require.ElementsMatch(t, tc.expect, output)
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
			"valid",
			[]model.Parameter{
				{
					Name:  "project",
					Value: "my-gcp-project",
				},
			},
			`{"project":"my-gcp-project"}`,
			"",
		},
		{
			"multi-valid",
			[]model.Parameter{
				{
					Name:  "a",
					Value: "b",
				},
				{
					Name:  "int",
					Value: float64(12),
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
			`{"a":"b","int":12,"map":{"bool":true,"num":401,"x":"y"},"project":"my-gcp-project"}`,
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ParametersToString(tc.input)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.Equal(t, tc.expect, output)
		})
	}
}
