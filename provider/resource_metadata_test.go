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
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/stretchr/testify/require"
)

// configurationResources are the configuration resources that support
// a description.
func configurationResources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"bindplane_configuration":    resourceConfiguration(),
		"bindplane_configuration_v2": resourceConfigurationV2(),
	}
}

func TestConfigurationSchemaDescription(t *testing.T) {
	for name, r := range configurationResources() {
		t.Run(name, func(t *testing.T) {
			s, ok := r.Schema["description"]
			require.True(t, ok, "missing description")
			require.Equal(t, schema.TypeString, s.Type)
			require.True(t, s.Optional)
			require.False(t, s.Required)
			require.False(t, s.ForceNew)
			require.Contains(t, s.Description, "configuration")
			require.NotContains(t, s.Description, "component")

			_, ok = r.Schema["display_name"]
			require.False(t, ok, "display_name is not supported on configurations")
		})
	}
}

func TestConfigurationDescriptionCreateUpdate(t *testing.T) {
	for name, r := range configurationResources() {
		t.Run(name, func(t *testing.T) {
			for _, rollout := range []bool{true, false} {
				fake, meta := newFakeBindplane(t)

				// Create
				d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
					"name":        "my-config",
					"platform":    "linux",
					"rollout":     rollout,
					"description": "collects host metrics",
				})
				require.NoError(t, r.Create(d, meta))
				require.Equal(t, "configuration-my-config", d.Id())
				require.Equal(t, "collects host metrics", fake.lastApplied().Metadata.Description)
				require.Equal(t, "collects host metrics", d.Get("description"))

				// Update the description
				d = schema.TestResourceDataRaw(t, r.Schema, map[string]any{
					"name":        "my-config",
					"platform":    "linux",
					"rollout":     rollout,
					"description": "new description",
				})
				d.SetId("configuration-my-config")
				require.NoError(t, r.Update(d, meta))
				require.Equal(t, "new description", fake.lastApplied().Metadata.Description)
				require.Equal(t, "new description", d.Get("description"))

				// Remove the description
				d = schema.TestResourceDataRaw(t, r.Schema, map[string]any{
					"name":     "my-config",
					"platform": "linux",
					"rollout":  rollout,
				})
				d.SetId("configuration-my-config")
				require.NoError(t, r.Update(d, meta))
				require.Equal(t, "", fake.lastApplied().Metadata.Description)
				require.Equal(t, "", d.Get("description"))

				if rollout {
					require.Equal(t, []string{"my-config", "my-config", "my-config"}, fake.rollouts)
				} else {
					require.Empty(t, fake.rollouts)
				}
			}
		})
	}
}

func TestConfigurationDescriptionRead(t *testing.T) {
	cases := []struct {
		name        string
		state       string
		server      string
		expectState string
	}{
		{
			name:        "set",
			state:       "collects host metrics",
			server:      "collects host metrics",
			expectState: "collects host metrics",
		},
		{
			name:        "unset",
			server:      "",
			expectState: "",
		},
		{
			name:        "changed-outside-terraform",
			state:       "collects host metrics",
			server:      "changed in the UI",
			expectState: "changed in the UI",
		},
		{
			// A description cleared outside of Terraform must clear
			// state so that the removal shows as a diff.
			name:        "removed-outside-terraform",
			state:       "collects host metrics",
			server:      "",
			expectState: "",
		},
	}

	for name, r := range configurationResources() {
		for _, tc := range cases {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				fake, meta := newFakeBindplane(t)
				fake.configurations["my-config"] = &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						Kind: model.KindConfiguration,
						Metadata: model.Metadata{
							ID:          "config-id",
							Name:        "my-config",
							Description: tc.server,
						},
					},
				}

				d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
					"name":        "my-config",
					"platform":    "linux",
					"rollout":     true,
					"description": tc.state,
				})
				d.SetId("config-id")

				require.NoError(t, r.Read(d, meta))
				require.Equal(t, "config-id", d.Id())
				require.Equal(t, tc.expectState, d.Get("description"))
			})
		}

		t.Run(name+"/id-mismatch", func(t *testing.T) {
			// A configuration with the same name, created outside of
			// Terraform, clears the ID.
			fake, meta := newFakeBindplane(t)
			fake.configurations["my-config"] = &model.Configuration{
				ResourceMeta: model.ResourceMeta{
					Kind: model.KindConfiguration,
					Metadata: model.Metadata{
						ID:          "other-id",
						Name:        "my-config",
						Description: "created elsewhere",
					},
				},
			}
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
				"name":        "my-config",
				"platform":    "linux",
				"rollout":     true,
				"description": "collects host metrics",
			})
			d.SetId("config-id")

			require.NoError(t, r.Read(d, meta))
			require.Equal(t, "", d.Id())
			require.Equal(t, "collects host metrics", d.Get("description"))
		})

		t.Run(name+"/server-error", func(t *testing.T) {
			_, meta := newFakeBindplane(t)
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
				"name":     fakeServerErrorName,
				"platform": "linux",
				"rollout":  true,
			})
			d.SetId("config-id")

			require.ErrorContains(t, r.Read(d, meta), "500")
		})

		t.Run(name+"/not-found", func(t *testing.T) {
			_, meta := newFakeBindplane(t)
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
				"name":        "my-config",
				"platform":    "linux",
				"rollout":     true,
				"description": "collects host metrics",
			})
			d.SetId("config-id")

			require.NoError(t, r.Read(d, meta))
			require.Equal(t, "", d.Id())
		})
	}
}

func TestProcessorBundleCreateMetadata(t *testing.T) {
	fake, meta := newFakeBindplane(t)
	r := resourceProcessorBundle()

	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"name":         "my-bundle",
		"rollout":      false,
		"display_name": "My Bundle",
		"description":  "batches and counts",
		"processor": []any{
			map[string]any{"name": "batch"},
		},
	})
	require.NoError(t, r.Create(d, meta))

	applied := fake.lastApplied()
	require.Equal(t, model.KindProcessor, applied.Kind)
	require.Equal(t, "My Bundle", applied.Metadata.DisplayName)
	require.Equal(t, "batches and counts", applied.Metadata.Description)

	require.NotEmpty(t, d.Id())
	require.Equal(t, "My Bundle", d.Get("display_name"))
	require.Equal(t, "batches and counts", d.Get("description"))
}

func TestProcessorBundleReadMetadata(t *testing.T) {
	cases := []struct {
		name     string
		state    map[string]any
		metadata model.Metadata
	}{
		{
			name:  "set",
			state: map[string]any{},
			metadata: model.Metadata{
				DisplayName: "My Bundle",
				Description: "batches and counts",
			},
		},
		{
			name:     "unset",
			state:    map[string]any{},
			metadata: model.Metadata{},
		},
		{
			name: "changed-outside-terraform",
			state: map[string]any{
				"display_name": "My Bundle",
				"description":  "batches and counts",
			},
			metadata: model.Metadata{
				DisplayName: "Renamed in the UI",
				Description: "changed in the UI",
			},
		},
		{
			// Values cleared outside of Terraform must clear state so
			// that the removal shows as a diff.
			name: "removed-outside-terraform",
			state: map[string]any{
				"display_name": "My Bundle",
				"description":  "batches and counts",
			},
			metadata: model.Metadata{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake, meta := newFakeBindplane(t)
			r := resourceProcessorBundle()

			md := tc.metadata
			md.ID = "bundle-id"
			md.Name = "my-bundle"
			fake.processors["my-bundle"] = newTestProcessorBundle(md)

			state := map[string]any{
				"name":    "my-bundle",
				"rollout": false,
			}
			for k, v := range tc.state {
				state[k] = v
			}
			d := schema.TestResourceDataRaw(t, r.Schema, state)
			d.SetId("bundle-id")

			require.NoError(t, r.Read(d, meta))
			require.Equal(t, "bundle-id", d.Id())
			require.Equal(t, tc.metadata.DisplayName, d.Get("display_name"))
			require.Equal(t, tc.metadata.Description, d.Get("description"))
			require.Equal(t, []any{map[string]any{"name": "batch"}}, d.Get("processor"))
		})
	}

	t.Run("server-error", func(t *testing.T) {
		_, meta := newFakeBindplane(t)
		r := resourceProcessorBundle()

		d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
			"name":    fakeServerErrorName,
			"rollout": false,
		})
		d.SetId("bundle-id")

		require.ErrorContains(t, r.Read(d, meta), "500")
	})

	t.Run("not-found", func(t *testing.T) {
		_, meta := newFakeBindplane(t)
		r := resourceProcessorBundle()

		d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
			"name":         "my-bundle",
			"rollout":      false,
			"display_name": "My Bundle",
		})
		d.SetId("bundle-id")

		require.NoError(t, r.Read(d, meta))
		require.Equal(t, "", d.Id())
	})
}

func TestProcessorBundleImportMetadata(t *testing.T) {
	fake, meta := newFakeBindplane(t)
	r := resourceProcessorBundle()

	fake.processors["my-bundle"] = newTestProcessorBundle(model.Metadata{
		ID:          "bundle-id",
		Name:        "my-bundle",
		DisplayName: "My Bundle",
		Description: "batches and counts",
	})

	// Terraform import sets the ID to the name given by the user.
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
	d.SetId("my-bundle")

	imported, err := r.Importer.StateContext(context.Background(), d, meta)
	require.NoError(t, err)
	require.Len(t, imported, 1)

	// Terraform reads the resource after import.
	require.NoError(t, r.Read(imported[0], meta))
	require.Equal(t, "bundle-id", imported[0].Id())
	require.Equal(t, "my-bundle", imported[0].Get("name"))
	require.Equal(t, "My Bundle", imported[0].Get("display_name"))
	require.Equal(t, "batches and counts", imported[0].Get("description"))

	t.Run("not-found", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
		d.SetId("missing")

		_, err := r.Importer.StateContext(context.Background(), d, meta)
		require.ErrorContains(t, err, "does not exist")
	})

	t.Run("server-error", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
		d.SetId(fakeServerErrorName)

		_, err := r.Importer.StateContext(context.Background(), d, meta)
		require.ErrorContains(t, err, "500")
	})
}

// newTestProcessorBundle returns a processor bundle with the given metadata
// and a single "batch" processor.
func newTestProcessorBundle(md model.Metadata) *model.Processor {
	p := &model.Processor{}
	p.Kind = model.KindProcessor
	p.Metadata = md
	p.Spec.Type = "processor_bundle"
	p.Spec.Processors = []model.ResourceConfiguration{{Name: "batch"}}
	return p
}
