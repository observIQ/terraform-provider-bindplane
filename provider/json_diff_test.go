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
	"testing"
)

func TestSuppressEquivalentJSONDiffs(t *testing.T) {
	tests := []struct {
		name     string
		old      string
		new      string
		expected bool
	}{
		{
			name:     "identical json",
			old:      `{"key":"value"}`,
			new:      `{"key":"value"}`,
			expected: true,
		},
		{
			name:     "whitespace differences",
			old:      `{"key":"value"}`,
			new:      `{"key": "value"}`,
			expected: true,
		},
		{
			name:     "different formatting with newlines",
			old:      `{"key":"value","nested":{"foo":"bar"}}`,
			new:      "{\n  \"key\": \"value\",\n  \"nested\": {\n    \"foo\": \"bar\"\n  }\n}",
			expected: true,
		},
		{
			name:     "different key order but same content",
			old:      `{"a":"1","b":"2"}`,
			new:      `{"b":"2","a":"1"}`,
			expected: true,
		},
		{
			name:     "array with same elements",
			old:      `[{"name":"test","value":"123"}]`,
			new:      `[{"name": "test", "value": "123"}]`,
			expected: true,
		},
		{
			name:     "different values",
			old:      `{"key":"value1"}`,
			new:      `{"key":"value2"}`,
			expected: false,
		},
		{
			name:     "different keys",
			old:      `{"key1":"value"}`,
			new:      `{"key2":"value"}`,
			expected: false,
		},
		{
			name:     "both empty strings",
			old:      "",
			new:      "",
			expected: true,
		},
		{
			name:     "one empty string",
			old:      "",
			new:      `{"key":"value"}`,
			expected: false,
		},
		{
			name:     "invalid json old",
			old:      `{invalid}`,
			new:      `{"key":"value"}`,
			expected: false,
		},
		{
			name:     "invalid json new",
			old:      `{"key":"value"}`,
			new:      `{invalid}`,
			expected: false,
		},
		{
			name:     "complex nested structure",
			old:      `[{"name":"param1","value":{"nested":true},"sensitive":false},{"name":"param2","value":"test"}]`,
			new:      "[\n  {\n    \"name\": \"param1\",\n    \"value\": {\n      \"nested\": true\n    },\n    \"sensitive\": false\n  },\n  {\n    \"name\": \"param2\",\n    \"value\": \"test\"\n  }\n]",
			expected: true,
		},
		{
			name:     "different array order",
			old:      `[{"name":"param1"},{"name":"param2"}]`,
			new:      `[{"name":"param2"},{"name":"param1"}]`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := suppressEquivalentJSONDiffs("parameters_json", tt.old, tt.new, nil)
			if result != tt.expected {
				t.Errorf("suppressEquivalentJSONDiffs() = %v, expected %v\nold: %s\nnew: %s",
					result, tt.expected, tt.old, tt.new)
			}
		})
	}
}
