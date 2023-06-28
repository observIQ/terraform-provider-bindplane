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

// Package resource provides functions for defining bindplane
// generic resources.

package parameter

// StringToParameter converts serialized json key values pairs
// to a list of BindPlane parameters.
// func StringToParameter(s string) ([]model.Parameter, error) {
// 	paramMap := make(map[string]any)
// 	if err := json.Unmarshal([]byte(s), &paramMap); err != nil {
// 		return nil, fmt.Errorf("failed to convert string parameters to map[string]any: %v", err)
// 	}

// 	parameters := []model.Parameter{}

// 	for k, v := range paramMap {
// 		parameters = append(parameters, model.Parameter{
// 			Name:  k,
// 			Value: v,
// 		})
// 	}

// 	return parameters, nil
// }

// // ParametersToSring converts a list of parameters to
// // serialized json key values pairs.
// func ParametersToString(params []model.Parameter) (string, error) {
// 	paramMap := make(map[string]any, len(params))
// 	for _, param := range params {
// 		paramMap[param.Name] = param.Value
// 	}

// 	b, err := json.Marshal(paramMap)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal parameters to json string %v", err)
// 	}

// 	return string(b), nil
// }
