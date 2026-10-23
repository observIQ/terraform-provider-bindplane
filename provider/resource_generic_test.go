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
