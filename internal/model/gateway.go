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

package model

import "strings"

// Gateway is a Bindplane gateway
type Gateway struct {
	ResourceMeta
	GatewaySpec `json:"spec"`
}

// GatewaySpec is the spec of a gateway
type GatewaySpec struct {
	// Configuration is the name of the configuration assigned to the
	// gateway's collectors, without a version
	Configuration string `json:"configuration,omitempty"`

	// Workload is set when the gateway is managed by the Bindplane
	// Kubernetes operator
	Workload *GatewayWorkloadRef `json:"workload,omitempty"`

	// LoadBalancerHostname is the hostname of the load balancer in front
	// of the gateway's collectors
	LoadBalancerHostname string `json:"loadBalancerHostname,omitempty"`
}

// GatewayWorkloadRef links an operator mode gateway to the Kubernetes
// workload that manages it. The provider only checks whether it is set.
type GatewayWorkloadRef struct{}

// GatewayMode is whether a gateway is managed manually or by the
// Bindplane Kubernetes operator
type GatewayMode string

// Gateway modes
const (
	GatewayModeManual   GatewayMode = "manual"
	GatewayModeOperator GatewayMode = "operator"
)

// Mode returns the mode of the gateway
func (g *Gateway) Mode() GatewayMode {
	if g.GatewaySpec.Workload != nil {
		return GatewayModeOperator
	}
	return GatewayModeManual
}

// TrimVersion removes the version from a resource key, e.g. "name:2"
// becomes "name". It does nothing if there is no version.
func TrimVersion(resourceKey string) string {
	name, _, _ := strings.Cut(resourceKey, ":")
	return name
}
