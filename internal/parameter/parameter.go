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

// Package parameter provides functions for marshalling and unmarshalling
//
//	resource parameters.
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
