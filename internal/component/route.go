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
	"fmt"
	"strings"

	"github.com/observiq/bindplane-op-enterprise/model"
)

const (
	// RouteTypeLogs routes logs
	RouteTypeLogs = "logs"

	// RouteTypeMetrics routes metrics
	RouteTypeMetrics = "metrics"

	// RouteTypeTraces routes traces
	RouteTypeTraces = "traces"

	// RoutePrefixProcessor is the prefix for processor routes
	RoutePrefixProcessor = "processors"

	// RoutePrefixDestination is the prefix for destination routes
	RoutePrefixDestination = "destinations"

	// RoutePrefixConnector is the prefix for connector routes
	RoutePrefixConnector = "connectors"
)

// ValidateRouteType returns an error if the route type is invalid
func ValidateRouteType(routeType string) error {
	switch routeType {
	case RouteTypeLogs, RouteTypeMetrics, RouteTypeTraces:
		return nil
	}
	return fmt.Errorf("invalid route type: %s", routeType)
}

// ValidateRouteComponents returns an error if the route components are invalid
func ValidateRouteComponents(components []model.ComponentPath) []error {
	var errs []error
	for _, c := range components {
		prefix := strings.Split(string(c), "/")[0]
		switch prefix {
		case RoutePrefixProcessor, RoutePrefixDestination, RoutePrefixConnector:
			continue
		default:
			errs = append(errs, fmt.Errorf("invalid route component: %s", c))
		}
	}
	return errs
}

// ParseRoutes reads routes from the state and returns
// them as a model.Routes object.
//
// This function expects to be passed the raw value from
// from (d *schema.ResourceData) Get Terraform SDK method.
// Schema validation ensures the type assertion used here
// will succeed.
func ParseRoutes(rawRoutes []any) (*model.Routes, error) {
	routes := &model.Routes{}

	for _, r := range rawRoutes {
		rawComponents := r.(map[string]any)["components"].([]any)
		components := []model.ComponentPath{}
		for _, c := range rawComponents {
			components = append(components, model.ComponentPath(c.(string)))
		}
		if err := ValidateRouteComponents(components); err != nil {
			return nil, fmt.Errorf("validate route components: %v", err)
		}

		// Schema validation ensures the type assertion used here
		id := r.(map[string]any)["route_id"].(string)

		telemetryType := r.(map[string]any)["telemetry_type"].(string)
		switch telemetryType {
		case RouteTypeLogs:
			routes.Logs = append(routes.Logs, model.Route{
				ID:         id,
				Components: components,
			})
		case RouteTypeMetrics:
			routes.Metrics = append(routes.Metrics, model.Route{
				ID:         id,
				Components: components,
			})
		case RouteTypeTraces:
			routes.Traces = append(routes.Traces, model.Route{
				ID:         id,
				Components: components,
			})
		}
	}

	return routes, nil
}

// RoutesToState takes a model.Routes object and returns a []map[string]any
// which is suitable for writing to the state.
//
// Returns nil, nil if routes is nil. It is up to the caller to ensure
// the returned value is not nil before attempting to read from it.
func RoutesToState(inRoutes *model.Routes) ([]map[string]any, error) {
	if inRoutes == nil {
		return nil, nil
	}

	stateRoutes := []map[string]any{}

	routeMap := map[string][]model.Route{
		RouteTypeLogs:    inRoutes.Logs,
		RouteTypeMetrics: inRoutes.Metrics,
		RouteTypeTraces:  inRoutes.Traces,
	}

	for routeType, routes := range routeMap {
		if len(routes) > 0 {
			r := []map[string]any{}
			for _, route := range routes {
				r = append(r, map[string]any{
					"route_id":       route.ID,
					"telemetry_type": routeType,
					"components":     route.Components,
				})
			}
			stateRoutes = append(stateRoutes, r...)
		}
	}

	return stateRoutes, nil
}
