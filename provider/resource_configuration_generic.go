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

// readRolloutOptions safely reads "rollout_options" from the resource data.
func readRolloutOptions(d *schema.ResourceData) (model.ResourceConfiguration, error) {
	rolloutOptionsRaw, ok := d.GetOk("rollout_options")
	if !ok || len(rolloutOptionsRaw.([]interface{})) == 0 {
		return model.ResourceConfiguration{}, nil
	}

	// Because d.GetOk returned a non nil value, we can assume that the
	// rollout_options list has at least one element due to the Terraform
	// framework's schema validation. Type assertion is safe in this case.

	rolloutOptions := rolloutOptionsRaw.([]interface{})[0].(map[string]interface{})
	resourceConfig := model.ResourceConfiguration{}

	if t, ok := rolloutOptions["type"].(string); ok {
		resourceConfig.Type = t
	}

	if parametersRaw, ok := rolloutOptions["parameters"]; ok {
		parametersList := parametersRaw.([]interface{})
		parameters := make([]model.Parameter, len(parametersList))
		for i, p := range parametersList {
			paramMap := p.(map[string]interface{})
			param := model.Parameter{}
			if name, ok := paramMap["name"].(string); ok {
				param.Name = name
			}
			if valueRaw, ok := paramMap["value"]; ok {
				valueList := valueRaw.([]interface{})
				values := make([]interface{}, len(valueList))
				for j, v := range valueList {
					values[j] = v.(map[string]interface{})
				}
				param.Value = values
			}
			parameters[i] = param
		}
		resourceConfig.Parameters = parameters
	}

	return resourceConfig, nil
}
