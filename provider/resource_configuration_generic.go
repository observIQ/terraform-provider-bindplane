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

package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op-enterprise/model"
)

// genericConfigurationDelete deletes configurations and raw configurations.
func genericConfigurationDelete(d *schema.ResourceData, meta any) error {
	return genericResourceDelete(model.KindConfiguration, d, meta)
}

func isValidPlatform(platform string) bool {
	// TODO(jsirianni): We should use a bindplane-op package to determine
	// valid platforms.
	const (
		platformWindows             = "windows"
		platformLinux               = "linux"
		platformMacOS               = "macos"
		platformK8sDaemonset        = "kubernetes-daemonset"
		platformK8sDeployment       = "kubernetes-deployment"
		platformGateway             = "kubernetes-gateway"
		platformOpenshiftDaemonset  = "openshift-daemonset"
		platformOpenshiftDeployment = "openshift-deployment"
	)
	switch platform {
	case platformWindows, platformLinux, platformMacOS,
		platformK8sDaemonset, platformK8sDeployment, platformGateway,
		platformOpenshiftDaemonset, platformOpenshiftDeployment:
		return true
	default:
		return false
	}
}
