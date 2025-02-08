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

	// RouteTypeLogsMetrics routes logs and metrics
	RouteTypeLogsMetrics = "logs+metrics"

	// RouteTypeLogsTraces routes logs and traces
	RouteTypeLogsTraces = "logs+traces"

	// RouteTypeMetricsTraces routes metrics and traces
	RouteTypeMetricsTraces = "metrics+traces"

	// RouteTypeLogsMetricsTraces routes logs, metrics, and traces
	RouteTypeLogsMetricsTraces = "logs+metrics+traces"

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
	case RouteTypeLogs, RouteTypeMetrics, RouteTypeTraces, RouteTypeLogsMetrics, RouteTypeLogsTraces, RouteTypeMetricsTraces, RouteTypeLogsMetricsTraces:
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
