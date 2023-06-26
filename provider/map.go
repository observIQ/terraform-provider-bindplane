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

import "fmt"

// stringMapFromTFMap takes a map[string]any and converts it to
// a map[string]string. Returns an error if any value is not a string.
func stringMapFromTFMap(m map[string]any) (map[string]string, error) {
	if m == nil {
		return nil, nil
	}

	strMap := make(map[string]string, len(m))

	for k, v := range m {
		switch v := v.(type) {
		case string:
			strMap[k] = v
		default:
			return nil, fmt.Errorf("expected value for key %s to be a string, got %T", k, v)
		}
	}

	return strMap, nil
}
