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
	"strings"
	"testing"

	"github.com/observiq/terraform-provider-bindplane/internal/model"
	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	routes := &model.Routes{
		Logs: []model.Route{{ID: "r1", Components: []model.ComponentPath{"destinations/dest-1"}}},
	}
	rollout := model.ResourceConfiguration{
		ParameterizedSpec: model.ParameterizedSpec{
			Type: "progressive",
			Parameters: []model.Parameter{
				{Name: "stages", Value: []any{map[string]any{"name": "stage"}}},
			},
		},
	}

	c, err := NewV2(
		WithName("my-config"),
		WithConnectorsByName([]ResourceConfig{
			{Name: "my-connector", RouteID: "conn-1", Routes: routes},
		}),
		WithProcessorGroups([]ResourceConfig{
			{RouteID: "pg-1", Processors: []string{"my-batch", "my-filter"}, Routes: routes},
		}),
		WithExtensionsByName([]ResourceConfig{
			{Name: "my-pprof"},
		}),
		WithRolloutOptions(rollout),
		WithMeasurementInterval("15m"),
	)
	require.NoError(t, err)

	require.Equal(t, "my-config", c.Metadata.Name)
	require.Equal(t, []model.ResourceConfiguration{
		{
			ID:   "conn-1",
			Name: "my-connector",
			ParameterizedSpec: model.ParameterizedSpec{
				Processors: []model.ResourceConfiguration{},
			},
			Routes: routes,
		},
	}, c.Spec.Connectors)
	require.Equal(t, []model.ResourceConfiguration{
		{
			ID: "pg-1",
			ParameterizedSpec: model.ParameterizedSpec{
				Processors: []model.ResourceConfiguration{
					{Name: "my-batch"},
					{Name: "my-filter"},
				},
			},
			Routes: routes,
		},
	}, c.Spec.Processors)
	require.Equal(t, []model.ResourceConfiguration{
		{
			Name: "my-pprof",
			ParameterizedSpec: model.ParameterizedSpec{
				Processors: []model.ResourceConfiguration{},
			},
		},
	}, c.Spec.Extensions)
	require.Equal(t, rollout, c.Spec.Rollout)
	require.Equal(t, "15m", c.Spec.MeasurementInterval)
}

func TestWithLabels(t *testing.T) {
	cases := []struct {
		name      string
		labels    map[string]string
		expectErr bool
	}{
		{"empty", map[string]string{}, false},
		{"simple", map[string]string{"env": "production"}, false},
		{"all valid characters", map[string]string{"Env_1-a.b": "Value_1-a.b"}, false},
		{"prefixed key", map[string]string{"bindplane.com/env": "production"}, false},
		{"empty value", map[string]string{"env": ""}, false},
		{"63 character name", map[string]string{strings.Repeat("a", 63): "v"}, false},
		{"63 character value", map[string]string{"env": strings.Repeat("a", 63)}, false},
		{"empty name", map[string]string{"": "v"}, true},
		{"name starts with dash", map[string]string{"-env": "v"}, true},
		{"name ends with dot", map[string]string{"env.": "v"}, true},
		{"name with space", map[string]string{"my env": "v"}, true},
		{"64 character name", map[string]string{strings.Repeat("a", 64): "v"}, true},
		{"empty prefix", map[string]string{"/env": "v"}, true},
		{"empty name after prefix", map[string]string{"bindplane.com/": "v"}, true},
		{"uppercase prefix", map[string]string{"Bindplane.com/env": "v"}, true},
		{"too many slashes", map[string]string{"a/b/c": "v"}, true},
		{"254 character prefix", map[string]string{strings.Repeat("a", 254) + "/env": "v"}, true},
		{"value starts with dash", map[string]string{"env": "-v"}, true},
		{"value with space", map[string]string{"env": "my value"}, true},
		{"64 character value", map[string]string{"env": strings.Repeat("a", 64)}, true},
		{"one invalid label", map[string]string{"env": "production", "bad key": "v"}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewV1(WithLabels(tc.labels))
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.labels, c.Metadata.Labels.AsMap())
		})
	}
}
