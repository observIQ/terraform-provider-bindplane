package component

import (
	"fmt"
	"strings"

	"github.com/observiq/bindplane-op-enterprise/model"
)

const (
	// Logs route type
	RouteTypeLogs = "logs"

	// Metrics route type
	RouteTypeMetrics = "metrics"

	// Traces route type
	RouteTypeTraces = "traces"

	// Logs + Metrics route type
	RouteTypeLogsMetrics = "logs+metrics"

	// Logs + Traces route type
	RouteTypeLogsTraces = "logs+traces"

	// Metrics + Traces route type
	RouteTypeMetricsTraces = "metrics+traces"

	// Logs + Metrics + Traces route type
	RouteTypeLogsMetricsTraces = "logs+metrics+traces"
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
		case "destinations", "processors", "connectors":
			continue
		default:
			errs = append(errs, fmt.Errorf("invalid route component: %s", c))
		}
	}
	return errs
}
