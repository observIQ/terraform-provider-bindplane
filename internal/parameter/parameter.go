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

// Package parameter provides functions for marshalling and unmarshalling resource parameters.
package parameter

import (
	"encoding/json"
	"fmt"

	"github.com/observiq/bindplane-op-enterprise/model"
)

// StringToParameter unmarshals serialized json parameters
// to a list of Bindplane parameters.
func StringToParameter(s string) ([]model.Parameter, error) {
	parameters := []model.Parameter{}
	if err := json.Unmarshal([]byte(s), &parameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters '%s': %w", s, err)
	}

	if err := validateParameters(parameters); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	return parameters, nil
}

// ParametersToString converts a list of parameters to
// serialized json key values pairs.
func ParametersToString(p []model.Parameter) (string, error) {
	if len(p) == 0 {
		return "", nil
	}

	paramBytes, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("failed to marshal parameters: %w", err)
	}

	return string(paramBytes), nil
}

// validateParameters validates the parameters for certain parameter types
func validateParameters(parameters []model.Parameter) error {
	for i, param := range parameters {
		switch param.Name {
		case "condition":
			if err := validateParametersForConditions(param, i); err != nil {
				return err
			}
		// future parameter validation can be added here
		default:
			continue
		}
	}
	return nil
}

// validateParametersForConditions validates parameters for malformed condition UI blocks
func validateParametersForConditions(param model.Parameter, paramIndex int) error {
	conditionBytes, err := json.Marshal(param.Value)
	if err != nil {
		return fmt.Errorf("parameter %d: failed to marshal condition: %w", paramIndex, err)
	}

	var condition Condition
	if err := json.Unmarshal(conditionBytes, &condition); err != nil {
		return fmt.Errorf("parameter %d: malformed condition structure: %w", paramIndex, err)
	}

	if err := ValidateOTTLConditionStatement(condition.UI); err != nil {
		return fmt.Errorf("parameter %d: invalid condition UI: %w", paramIndex, err)
	}

	return nil
}
