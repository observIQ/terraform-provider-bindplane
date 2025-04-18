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
					"route_id":       "logs-otlp",
					"components":     []any{"destinations/otlp"},
					"telemetry_type": "logs",
				},
			},
			&model.Routes{
				Logs: []model.Route{
					{
						ID:         "logs-otlp",
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
					"route_id":       "logs",
					"components":     []any{"destinations/loki", "connectors/router"},
					"telemetry_type": "logs",
				},
				map[string]any{
					"route_id":       "metrics",
					"components":     []any{"processors/batcher"},
					"telemetry_type": "metrics",
				},
				map[string]any{
					"route_id":       "traces",
					"components":     []any{"destinations/jaeger"},
					"telemetry_type": "traces",
				},
			},
			&model.Routes{
				Logs: []model.Route{
					{
						ID:         "logs",
						Components: []model.ComponentPath{"destinations/loki", "connectors/router"},
					},
				},
				Metrics: []model.Route{
					{
						ID:         "metrics",
						Components: []model.ComponentPath{"processors/batcher"},
					},
				},
				Traces: []model.Route{
					{
						ID:         "traces",
						Components: []model.ComponentPath{"destinations/jaeger"},
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
					"telemetry_type": "traces",
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
func TestRoutesToState(t *testing.T) {
	cases := []struct {
		name     string
		routes   *model.Routes
		expected []map[string]any
	}{
		{
			"nil routes",
			nil,
			nil,
		},
		{
			"empty routes",
			&model.Routes{},
			[]map[string]any{},
		},
		{
			"logs route",
			&model.Routes{
				Logs: []model.Route{
					{
						ID:         "loki",
						Components: []model.ComponentPath{"destinations/loki"},
					},
				},
			},
			[]map[string]any{
				{
					"route_id":       "loki",
					"telemetry_type": "logs",
					"components":     []model.ComponentPath{"destinations/loki"},
				},
			},
		},
		{
			"metrics route",
			&model.Routes{
				Metrics: []model.Route{
					{
						ID:         "data",
						Components: []model.ComponentPath{"processors/batcher"},
					},
				},
			},
			[]map[string]any{
				{
					"route_id":       "data",
					"telemetry_type": "metrics",
					"components":     []model.ComponentPath{"processors/batcher"},
				},
			},
		},
		{
			"traces route",
			&model.Routes{
				Traces: []model.Route{
					{
						ID:         "jaeger",
						Components: []model.ComponentPath{"destinations/jaeger"},
					},
				},
			},
			[]map[string]any{
				{
					"route_id":       "jaeger",
					"telemetry_type": "traces",
					"components":     []model.ComponentPath{"destinations/jaeger"},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state, err := RoutesToState(tc.routes)
			require.NoError(t, err)
			require.Equal(t, tc.expected, state)
		})
	}
}
