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

package maputil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringMapFromTFMap(t *testing.T) {
	cases := []struct {
		name   string
		input  map[string]any
		expect map[string]string
	}{
		{
			"string values",
			map[string]any{
				"x":    "y",
				"user": "name",
			},
			map[string]string{
				"x":    "y",
				"user": "name",
			},
		},
		{
			"int values",
			map[string]any{
				"x": 2,
				"y": 77,
			},
			nil,
		},
		{
			"hybrid values",
			map[string]any{
				"y": "test",
				"x": 2,
				"z": map[string]any{},
			},
			nil,
		},
		{
			"nil",
			nil,
			nil,
		},
	}

	for _, tc := range cases {
		output, err := StringMapFromTFMap(tc.input)
		if tc.expect == nil {
			require.Error(t, err, "expected an error when non string values are used")
			return
		}
		require.NoError(t, err)
		require.Equal(t, tc.expect, output)
	}
}
