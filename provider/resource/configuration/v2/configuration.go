// Package v2 provides schemas for configuration v2 resources.
package v2

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/terraform-provider-bindplane/internal/component"
)

var RouteSchema *schema.Schema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	ForceNew: false,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"telemetry_type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The telemetry type to route. Valid route types include 'logs', 'metrics', or 'traces' 'logs+metrics', 'logs+traces', 'metrics+traces', 'logs+metrics+traces'.",
				ValidateFunc: func(val any, _ string) (warns []string, errs []error) {
					telemetryType := val.(string)
					if err := component.ValidateRouteType(telemetryType); err != nil {
						errs = append(errs, err)
					}
					return
				},
				Default: component.RouteTypeLogsMetricsTraces,
			},
			"components": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    false,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of component names to route.",
			},
		},
	},
	Description: "Route telemetry to specific components.",
}
