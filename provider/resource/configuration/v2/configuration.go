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

// Package v2 provides schemas for configuration v2 resources.
package v2

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/terraform-provider-bindplane/internal/component"
)

// RouteSchema defines the schema for a route.
var RouteSchema *schema.Schema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	ForceNew: false,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"route_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "The unique identifier for the route.",
			},
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
