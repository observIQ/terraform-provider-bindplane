// Copyright observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package component

import (
	"errors"
	"fmt"
	"testing"

	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/stretchr/testify/require"
)

func TestValidateRouteType(t *testing.T) {
	cases := []struct {
		routeType string
		expected  error
	}{
		{"logs", nil},
		{"metrics", nil},
		{"traces", nil},
		{"logs+metrics", nil},
		{"logs+traces", nil},
		{"metrics+traces", nil},
		{"logs+metrics+traces", nil},
		{"invalid", errors.New("invalid route type: invalid")},
	}

	for _, c := range cases {
		err := ValidateRouteType(c.routeType)
		if c.expected != nil {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestValidateRouteComponents(t *testing.T) {
	cases := []struct {
		components []model.ComponentPath
		expected   []error
	}{
		{[]model.ComponentPath{"destinations", "processors", "connectors"}, nil},
		{[]model.ComponentPath{"destinations", "processors", "invalid"}, []error{fmt.Errorf("invalid route component: invalid")}},
	}

	for _, c := range cases {
		errs := ValidateRouteComponents(c.components)
		if len(c.expected) == 0 {
			require.Equal(t, c.expected, errs)
		}
		require.Len(t, errs, len(c.expected))
	}
}

func TestParseRoutes(t *testing.T) {
	cases := []struct {
		name      string
		rawRoutes []any
		expected  *model.Routes
		err       bool
	}{
		{
			"simple route",
			[]any{
				map[string]any{
					"components":     []any{"destinations/otlp"},
					"telemetry_type": "logs+metrics+traces",
				},
			},
			&model.Routes{
				LogsMetricsTraces: []model.Route{
					{
						Components: []model.ComponentPath{"destinations/otlp"},
					},
				},
			},
			false,
		},
		{
			"all valid route types",
			[]any{
				map[string]any{
					"components":     []any{"destinations/loki", "connectors/router"},
					"telemetry_type": "logs",
				},
				map[string]any{
					"components":     []any{"processors/batcher"},
					"telemetry_type": "metrics",
				},
				map[string]any{
					"components":     []any{"destinations/jaeger"},
					"telemetry_type": "traces",
				},
				map[string]any{
					"components":     []any{"destinations/otlp", "processors/batcher"},
					"telemetry_type": "logs+metrics",
				},
				map[string]any{
					"components":     []any{"destinations/otlp", "processors/batcher"},
					"telemetry_type": "logs+traces",
				},
				map[string]any{
					"components":     []any{"destinations/otlp", "processors/batcher"},
					"telemetry_type": "metrics+traces",
				},
				map[string]any{
					"components":     []any{"destinations/otlp", "processors/batcher", "connectors/router"},
					"telemetry_type": "logs+metrics+traces",
				},
			},
			&model.Routes{
				Logs: []model.Route{
					{
						Components: []model.ComponentPath{"destinations/loki", "connectors/router"},
					},
				},
				Metrics: []model.Route{
					{
						Components: []model.ComponentPath{"processors/batcher"},
					},
				},
				Traces: []model.Route{
					{
						Components: []model.ComponentPath{"destinations/jaeger"},
					},
				},
				LogsMetrics: []model.Route{
					{
						Components: []model.ComponentPath{"destinations/otlp", "processors/batcher"},
					},
				},
				LogsTraces: []model.Route{
					{
						Components: []model.ComponentPath{"destinations/otlp", "processors/batcher"},
					},
				},
				MetricsTraces: []model.Route{
					{
						Components: []model.ComponentPath{"destinations/otlp", "processors/batcher"},
					},
				},
				LogsMetricsTraces: []model.Route{
					{
						Components: []model.ComponentPath{"destinations/otlp", "processors/batcher", "connectors/router"},
					},
				},
			},
			false,
		},
		{
			"invalid route prefix",
			[]any{
				map[string]any{
					"components":     []any{"dest/otlp"},
					"telemetry_type": "logs+metrics+traces",
				},
			},
			nil,
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			routes, err := ParseRoutes(tc.rawRoutes)
			if tc.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expected, routes)
		})
	}
}
