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
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// suppressEquivalentJSONDiffs compares two JSON strings semantically,
// ignoring whitespace and formatting differences. This prevents Terraform
// from detecting spurious changes when the API returns JSON with different
// formatting than what was stored in state.
func suppressEquivalentJSONDiffs(_, old, new string, _ *schema.ResourceData) bool {
	// If both values are empty, consider them equivalent
	if old == "" && new == "" {
		return true
	}

	// If only one is empty, they are different
	if old == "" || new == "" {
		return false
	}

	// Unmarshal both JSON strings into generic interfaces
	var oldData, newData any

	if err := json.Unmarshal([]byte(old), &oldData); err != nil {
		// If old value is not valid JSON, fall back to string comparison
		return old == new
	}

	if err := json.Unmarshal([]byte(new), &newData); err != nil {
		// If new value is not valid JSON, fall back to string comparison
		return old == new
	}

	// Compare the unmarshaled data structures
	return reflect.DeepEqual(oldData, newData)
}
