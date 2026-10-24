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

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLabelsFromMap(t *testing.T) {
	cases := []struct {
		name      string
		key       string
		value     string
		expectErr string
	}{
		{"name", "env", "test", ""},
		{"name with symbols", "A.b-c_9", "test", ""},
		{"single character name", "a", "test", ""},
		{"max length name", strings.Repeat("a", 63), "test", ""},
		{"prefix", "bindplane.com/env", "test", ""},
		{"max length prefix", strings.Repeat("a", 253) + "/env", "test", ""},
		{"empty value", "env", "", ""},
		{"value with symbols", "env", "A.b-c_9", ""},
		{"max length value", "env", strings.Repeat("a", 63), ""},
		{"empty name", "", "test", "is not a valid label name: name part must be non-empty"},
		{"empty name with prefix", "bindplane.com/", "test", "is not a valid label name: name part must be non-empty"},
		{"name too long", strings.Repeat("a", 64), "test", "is not a valid label name: name part must be no more than 63 characters"},
		{"name with space", "bad key", "test", "is not a valid label name: name part must consist of alphanumeric characters"},
		{"name starts with symbol", "-env", "test", "is not a valid label name: name part must consist of alphanumeric characters"},
		{"name ends with symbol", "env.", "test", "is not a valid label name: name part must consist of alphanumeric characters"},
		{"empty prefix", "/env", "test", "is not a valid label name: prefix part must be non-empty"},
		{"prefix too long", strings.Repeat("a", 254) + "/env", "test", "is not a valid label name: prefix part must be no more than 253 characters"},
		{"prefix starts with uppercase", "Bindplane.com/env", "test", "is not a valid label name: prefix part must be a lowercase DNS subdomain"},
		{"prefix contains uppercase", "bindPlane.com/env", "test", "is not a valid label name: prefix part must be a lowercase DNS subdomain"},
		{"prefix starts with symbol", "-bindplane.com/env", "test", "is not a valid label name: prefix part must be a lowercase DNS subdomain"},
		{"prefix with empty subdomain", "bindplane..com/env", "test", "is not a valid label name: prefix part must be a lowercase DNS subdomain"},
		{"prefix with underscore", "bind_plane.com/env", "test", "is not a valid label name: prefix part must be a lowercase DNS subdomain"},
		{"multiple prefixes", "a/b/c", "test", "is not a valid label name: must be a name with an optional DNS subdomain prefix"},
		{"value too long", "env", strings.Repeat("a", 64), "is not a valid label value: must be no more than 63 characters"},
		{"value with space", "env", "bad value", "is not a valid label value: must be empty or consist of alphanumeric characters"},
		{"value starts with symbol", "env", "-test", "is not a valid label value: must be empty or consist of alphanumeric characters"},
		{"value ends with symbol", "env", "test_", "is not a valid label value: must be empty or consist of alphanumeric characters"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			labels, err := LabelsFromMap(map[string]string{tc.key: tc.value})
			if tc.expectErr != "" {
				require.ErrorContains(t, err, tc.expectErr)
				require.Empty(t, labels.AsMap(), "invalid labels are dropped")
				return
			}
			require.NoError(t, err)
			require.Equal(t, map[string]string{tc.key: tc.value}, labels.AsMap())
		})
	}
}

func TestLabelsFromMap_Multiple(t *testing.T) {
	labels, err := LabelsFromMap(map[string]string{
		"env":       "test",
		"platform":  "linux",
		"bad key":   "value",
		"bad-value": "-value",
	})
	require.Equal(t, map[string]string{"env": "test", "platform": "linux"}, labels.AsMap(), "valid labels are kept")
	require.ErrorContains(t, err, "bad key is not a valid label name")
	require.ErrorContains(t, err, "-value is not a valid label value")
}

func TestLabelsFromValidatedMap(t *testing.T) {
	// Labels are not validated
	m := map[string]string{"bad key": "-value"}
	require.Equal(t, m, LabelsFromValidatedMap(m).AsMap())

	require.Nil(t, LabelsFromValidatedMap(nil).AsMap())
}

func TestLabels_JSON(t *testing.T) {
	cases := []struct {
		name   string
		labels Labels
		expect string
	}{
		{"nil", Labels{}, `{}`},
		{"empty", LabelsFromValidatedMap(map[string]string{}), `{}`},
		{"labels", LabelsFromValidatedMap(map[string]string{"env": "test", "platform": "linux"}), `{"env":"test","platform":"linux"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Labels are marshaled the same way whether or not they are addressable
			byValue, err := json.Marshal(Metadata{Name: "name", Labels: tc.labels})
			require.NoError(t, err)
			byPointer, err := json.Marshal(&Metadata{Name: "name", Labels: tc.labels})
			require.NoError(t, err)

			expect := `{"id":"","name":"name","labels":` + tc.expect + `}`
			require.JSONEq(t, expect, string(byValue))
			require.JSONEq(t, expect, string(byPointer))

			var m Metadata
			require.NoError(t, json.Unmarshal(byValue, &m))
			require.Equal(t, len(tc.labels.AsMap()), len(m.Labels.AsMap()))
			for k, v := range tc.labels.AsMap() {
				require.Equal(t, v, m.Labels.AsMap()[k])
			}
		})
	}

	t.Run("null", func(t *testing.T) {
		var m Metadata
		require.NoError(t, json.Unmarshal([]byte(`{"name":"name","labels":null}`), &m))
		require.Nil(t, m.Labels.AsMap())
	})

	t.Run("invalid", func(t *testing.T) {
		var m Metadata
		require.Error(t, json.Unmarshal([]byte(`{"name":"name","labels":["env=test"]}`), &m))
	})
}
