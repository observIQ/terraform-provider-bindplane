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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/stretchr/testify/require"
)

func TestRemovedParameters(t *testing.T) {
	cases := []struct {
		name       string
		old        string
		newVal     string
		expect     []model.Parameter
		expectedOK bool
	}{
		{
			name:       "both-empty",
			old:        "",
			newVal:     "",
			expect:     nil,
			expectedOK: true,
		},
		{
			name:       "no-removals",
			old:        `[{"name":"project","value":"my-project"}]`,
			newVal:     `[{"name":"project","value":"my-project"}]`,
			expect:     nil,
			expectedOK: true,
		},
		{
			name:   "removed-from-state",
			old:    `[{"name":"project","value":"my-project"},{"name":"timeout","value":"12s"},{"name":"insecure","value":false}]`,
			newVal: `[{"name":"project","value":"my-project"}]`,
			expect: []model.Parameter{
				{Name: "timeout", Value: "12s"},
				{Name: "insecure", Value: false},
			},
			expectedOK: true,
		},
		{
			name:   "config-removed-entirely",
			old:    `[{"name":"timeout","value":"12s"}]`,
			newVal: "",
			expect: []model.Parameter{
				{Name: "timeout", Value: "12s"},
			},
			expectedOK: true,
		},
		{
			name:   "config-order-differs",
			old:    `[{"name":"timeout","value":"12s"},{"name":"project","value":"my-project"},{"name":"insecure","value":false}]`,
			newVal: `[{"name":"project","value":"my-project"},{"name":"timeout","value":"12s"}]`,
			expect: []model.Parameter{
				{Name: "insecure", Value: false},
			},
			expectedOK: true,
		},
		{
			name:   "numbers-compare-by-value",
			old:    `[{"name":"interval","value":60.0},{"name":"timeout","value":"12s"}]`,
			newVal: `[{"name":"interval","value":60}]`,
			expect: []model.Parameter{
				{Name: "timeout", Value: "12s"},
			},
			expectedOK: true,
		},
		{
			name:       "config-changes-a-value",
			old:        `[{"name":"project","value":"old-project"},{"name":"timeout","value":"12s"}]`,
			newVal:     `[{"name":"project","value":"my-project"}]`,
			expectedOK: false,
		},
		{
			name:       "config-changes-sensitive",
			old:        `[{"name":"token","value":"abc"},{"name":"timeout","value":"12s"}]`,
			newVal:     `[{"name":"token","value":"abc","sensitive":true}]`,
			expectedOK: false,
		},
		{
			name:       "config-adds-a-parameter",
			old:        `[{"name":"timeout","value":"12s"}]`,
			newVal:     `[{"name":"project","value":"my-project"}]`,
			expectedOK: false,
		},
		{
			name:       "invalid-old",
			old:        `not json`,
			newVal:     `[]`,
			expectedOK: false,
		},
		{
			name:       "invalid-new",
			old:        `[]`,
			newVal:     `not json`,
			expectedOK: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			removed, ok := removedParameters(tc.old, tc.newVal)
			require.Equal(t, tc.expectedOK, ok)
			if tc.expectedOK {
				require.Equal(t, tc.expect, removed)
			}
		})
	}
}

func TestParametersAreDefaults(t *testing.T) {
	defs := []model.ParameterDefinition{
		{Name: "timeout", Default: "12s"},
		{Name: "interval", Default: 60},
		{Name: "insecure", Default: false},
		{Name: "tags", Default: []any{"a", "b"}},
		{Name: "headers", Default: map[string]any{"x": "y"}},
		{Name: "credentials", Default: ""},
		{Name: "project"},
	}

	cases := []struct {
		name   string
		params []model.Parameter
		expect bool
	}{
		{
			name:   "no-params",
			params: nil,
			expect: true,
		},
		{
			name: "all-defaults",
			params: []model.Parameter{
				{Name: "timeout", Value: "12s"},
				// Decoded from JSON, numbers are float64.
				{Name: "interval", Value: float64(60)},
				{Name: "insecure", Value: false},
				{Name: "tags", Value: []any{"a", "b"}},
				{Name: "headers", Value: map[string]any{"x": "y"}},
				{Name: "credentials", Value: "", Sensitive: true},
			},
			expect: true,
		},
		{
			name: "non-default-value",
			params: []model.Parameter{
				{Name: "timeout", Value: "12s"},
				{Name: "interval", Value: float64(30)},
			},
			expect: false,
		},
		{
			name: "list-order-matters",
			params: []model.Parameter{
				{Name: "tags", Value: []any{"b", "a"}},
			},
			expect: false,
		},
		{
			// Bindplane masks sensitive values that are set, so the
			// value cannot be confirmed to be the default.
			name: "masked-sensitive-value",
			params: []model.Parameter{
				{Name: "credentials", Value: "(sensitive)", Sensitive: true},
			},
			expect: false,
		},
		{
			// Parameters without a default are not saved by Bindplane
			// when unset, so removing one is a real change.
			name: "no-default",
			params: []model.Parameter{
				{Name: "project", Value: "my-project"},
			},
			expect: false,
		},
		{
			name: "unknown-parameter",
			params: []model.Parameter{
				{Name: "removed_in_new_version", Value: "x"},
			},
			expect: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expect, parametersAreDefaults(tc.params, defs))
		})
	}
}

func TestSuppressDefaultParameterDiffsFallback(t *testing.T) {
	noMeta := func() any { return nil }
	suppress := suppressDefaultParameterDiffs(model.KindSource, noMeta)

	state := `[{"name":"project","value":"my-project"},{"name":"timeout","value":"12s"}]`
	config := `[{"name":"project","value":"my-project"}]`

	t.Run("equivalent-json", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceSource().Schema, map[string]any{})
		d.SetId("id")
		require.True(t, suppress("parameters_json", `[{"name":"a","value":1}]`, `[ {"name": "a", "value": 1} ]`, d))
	})

	t.Run("new-resource", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceSource().Schema, map[string]any{})
		require.False(t, suppress("parameters_json", state, config, d))
	})

	t.Run("no-client", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceSource().Schema, map[string]any{})
		d.SetId("id")
		require.False(t, suppress("parameters_json", state, config, d))
	})
}

func TestProviderParameterizedResources(t *testing.T) {
	p := Provider()
	for name := range parameterizedResources {
		t.Run(name, func(t *testing.T) {
			r, ok := p.ResourcesMap[name]
			require.True(t, ok, "resource %s is not registered", name)
			s, ok := r.Schema["parameters_json"]
			require.True(t, ok, "resource %s has no parameters_json", name)
			require.NotNil(t, s.DiffSuppressFunc)
		})
	}
}
