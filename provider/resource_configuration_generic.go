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
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op-enterprise/model"
)

var advancedSchema = &schema.Schema{
	Type:     schema.TypeList,
	Optional: true,
	MaxItems: 1,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"metrics": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"port": {
							Type:     schema.TypeInt,
							Optional: true,
							ValidateFunc: func(val any, _ string) (warns []string, errs []error) {
								port := val.(int)
								if port < 1 || port > 65535 {
									errs = append(errs, fmt.Errorf("%d is not a valid TCP port", port))
								}
								return
							},
							Description: "The advanced metrics port for the configuration.",
						},
						"level": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: func(val any, _ string) (warns []string, errs []error) {
								level := val.(string)
								if level != "normal" && level != "detailed" {
									errs = append(errs, fmt.Errorf("%s is not a valid advanced metrics level", level))
								}
								return
							},
							Description: "The advanced metrics level for the configuration.",
						},
					},
				},
			},
		},
	},
}

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

// extractAdvancedParameters extracts advanced parameters from the resource data.
func extractAdvancedParameters(d *schema.ResourceData) ([]model.Parameter, error) {
	advancedParameters := []model.Parameter{}
	if adv, ok := d.GetOk("advanced"); ok {
		advList := adv.([]any)
		if len(advList) > 0 {
			advMap := advList[0].(map[string]any)
			if metrics, ok := advMap["metrics"]; ok {
				metricsList := metrics.([]any)
				if len(metricsList) > 0 {
					metricsMap := metricsList[0].(map[string]any)
					if port, ok := metricsMap["port"]; ok {
						advancedParameters = append(advancedParameters, model.Parameter{Name: "telemetryPort", Value: port})
					}
					if level, ok := metricsMap["level"]; ok {
						advancedParameters = append(advancedParameters, model.Parameter{Name: "telemetryLevel", Value: level})
					}
				}
			}
		}
	}
	return advancedParameters, nil
}

// setAdvancedMetricsInState extracts advanced metrics from spec.parameters and sets them in the state.
func setAdvancedMetricsInState(d *schema.ResourceData, config *model.Configuration) error {
	advancedMetrics := map[string]any{}
	for _, param := range config.Spec.Parameters {
		if param.Name == "telemetryPort" {
			switch v := param.Value.(type) {
			case int:
				advancedMetrics["port"] = v
			case int64:
				advancedMetrics["port"] = int(v)
			case float64:
				advancedMetrics["port"] = int(v)
			}
		}
		if param.Name == "telemetryLevel" {
			if level, ok := param.Value.(string); ok {
				advancedMetrics["level"] = level
			}
		}
	}

	// Set advanced metrics in state
	if len(advancedMetrics) > 0 {
		if err := d.Set("advanced", []any{map[string]any{"metrics": []any{advancedMetrics}}}); err != nil {
			return fmt.Errorf("error setting advanced metrics: %s", err)
		}
	}
	return nil
}
