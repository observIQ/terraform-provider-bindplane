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
	"github.com/observiq/terraform-provider-bindplane/internal/model"
	"github.com/stretchr/testify/require"
)

func TestReconcileParameters(t *testing.T) {
	// Parameters as returned by the Bindplane API: every parameter
	// supported by the resource type, with defaults resolved.
	incoming := []model.Parameter{
		{Name: "telemetry_types", Value: []any{"Logs"}},
		{Name: "project", Value: "my-project"},
		{Name: "auth_type", Value: "auto"},
		{Name: "credentials", Value: "(sensitive)", Sensitive: true},
		{Name: "timeout", Value: "12s"},
		{Name: "insecure", Value: false},
	}

	cases := []struct {
		name     string
		state    []model.Parameter
		incoming []model.Parameter
		expect   []model.Parameter
	}{
		{
			name:     "no-state-params",
			state:    nil,
			incoming: incoming,
			expect:   []model.Parameter{},
		},
		{
			name:     "no-incoming-params",
			state:    []model.Parameter{{Name: "project", Value: "my-project"}},
			incoming: nil,
			expect:   []model.Parameter{},
		},
		{
			name: "subset-of-params",
			state: []model.Parameter{
				{Name: "telemetry_types", Value: []any{"Logs"}},
				{Name: "project", Value: "my-project"},
			},
			incoming: incoming,
			expect: []model.Parameter{
				{Name: "telemetry_types", Value: []any{"Logs"}},
				{Name: "project", Value: "my-project"},
			},
		},
		{
			name: "drift-on-configured-param",
			state: []model.Parameter{
				{Name: "project", Value: "old-project"},
			},
			incoming: incoming,
			expect: []model.Parameter{
				{Name: "project", Value: "my-project"},
			},
		},
		{
			name: "state-order-preserved",
			state: []model.Parameter{
				{Name: "timeout", Value: "12s"},
				{Name: "project", Value: "my-project"},
			},
			incoming: incoming,
			expect: []model.Parameter{
				{Name: "timeout", Value: "12s"},
				{Name: "project", Value: "my-project"},
			},
		},
		{
			name: "param-missing-from-api",
			state: []model.Parameter{
				{Name: "project", Value: "my-project"},
				{Name: "removed", Value: "value"},
			},
			incoming: incoming,
			expect: []model.Parameter{
				{Name: "project", Value: "my-project"},
			},
		},
		{
			name: "sensitive-value-from-state",
			state: []model.Parameter{
				{Name: "credentials", Value: "secret"},
			},
			incoming: incoming,
			expect: []model.Parameter{
				{Name: "credentials", Value: "secret"},
			},
		},
		{
			name: "sensitive-flag-from-state",
			state: []model.Parameter{
				{Name: "credentials", Value: "secret", Sensitive: true},
			},
			incoming: incoming,
			expect: []model.Parameter{
				{Name: "credentials", Value: "secret", Sensitive: true},
			},
		},
		{
			name:     "unconfigured-sensitive-param-dropped",
			state:    []model.Parameter{{Name: "auth_type", Value: "auto"}},
			incoming: incoming,
			expect: []model.Parameter{
				{Name: "auth_type", Value: "auto"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := reconcileParameters(tc.state, tc.incoming)
			require.Equal(t, tc.expect, output)
		})
	}
}

func TestImportParameters(t *testing.T) {
	cases := []struct {
		name     string
		incoming []model.Parameter
		expect   []model.Parameter
	}{
		{
			name:     "no-params",
			incoming: nil,
			expect:   []model.Parameter{},
		},
		{
			name: "all-params",
			incoming: []model.Parameter{
				{Name: "project", Value: "my-project"},
				{Name: "timeout", Value: "12s"},
			},
			expect: []model.Parameter{
				{Name: "project", Value: "my-project"},
				{Name: "timeout", Value: "12s"},
			},
		},
		{
			name: "sensitive-params-dropped",
			incoming: []model.Parameter{
				{Name: "project", Value: "my-project"},
				{Name: "credentials", Value: "(sensitive)", Sensitive: true},
			},
			expect: []model.Parameter{
				{Name: "project", Value: "my-project"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := importParameters(tc.incoming)
			require.Equal(t, tc.expect, output)
		})
	}
}

func TestSetResourceMetadata(t *testing.T) {
	cases := []struct {
		name     string
		state    map[string]any
		metadata model.Metadata
	}{
		{
			name:     "no-metadata",
			state:    map[string]any{},
			metadata: model.Metadata{},
		},
		{
			name:  "set-metadata",
			state: map[string]any{},
			metadata: model.Metadata{
				DisplayName: "My Host",
				Description: "collects host metrics",
			},
		},
		{
			name: "drift-on-metadata",
			state: map[string]any{
				"display_name": "Old Host",
				"description":  "old description",
			},
			metadata: model.Metadata{
				DisplayName: "My Host",
				Description: "collects host metrics",
			},
		},
		{
			// Metadata removed outside of Terraform must clear state
			// so that the removal shows as a diff.
			name: "metadata-removed",
			state: map[string]any{
				"display_name": "My Host",
				"description":  "collects host metrics",
			},
			metadata: model.Metadata{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceSource().Schema, tc.state)
			require.NoError(t, setResourceMetadata(d, tc.metadata))
			require.Equal(t, tc.metadata.DisplayName, d.Get("display_name"))
			require.Equal(t, tc.metadata.Description, d.Get("description"))
		})
	}
}

func TestGenericSchemaMetadata(t *testing.T) {
	resources := map[string]*schema.Resource{
		"bindplane_source":           resourceSource(),
		"bindplane_destination":      resourceDestination(),
		"bindplane_processor":        resourceProcessor(),
		"bindplane_processor_bundle": resourceProcessorBundle(),
		"bindplane_extension":        resourceExtension(),
		"bindplane_connector":        resourceConnector(),
	}

	for name, r := range resources {
		t.Run(name, func(t *testing.T) {
			for _, key := range []string{"display_name", "description"} {
				s, ok := r.Schema[key]
				require.True(t, ok, "missing %s", key)
				require.Equal(t, schema.TypeString, s.Type)
				require.True(t, s.Optional)
				require.False(t, s.ForceNew)
			}
		})
	}
}
