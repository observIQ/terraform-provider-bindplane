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
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/terraform-provider-bindplane/internal/model"
	"github.com/stretchr/testify/require"
)

func newGatewayData(t *testing.T, raw map[string]any) *schema.ResourceData {
	return schema.TestResourceDataRaw(t, resourceGateway().Schema, raw)
}

func testGatewayResource(id, name string, spec map[string]any) *model.AnyResource {
	return &model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       model.KindGateway,
			Metadata: model.Metadata{
				ID:   id,
				Name: name,
			},
		},
		Spec: spec,
	}
}

func testConfigurationResource(name string, labels map[string]string) *model.AnyResource {
	return &model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       model.KindConfiguration,
			Metadata: model.Metadata{
				ID:     name,
				Name:   name,
				Labels: model.LabelsFromValidatedMap(labels),
			},
		},
		Spec: map[string]any{},
	}
}

func TestResourceGatewaySchema(t *testing.T) {
	s := resourceGateway().Schema

	require.True(t, s["name"].Required)
	require.True(t, s["name"].ForceNew)

	for _, key := range []string{"display_name", "description", "configuration", "load_balancer_hostname"} {
		require.True(t, s[key].Optional, key)
		require.False(t, s[key].ForceNew, key)
		require.False(t, s[key].Computed, key)
	}
	require.Len(t, s, 5)

	r := resourceGateway()
	require.NotNil(t, r.Importer)
	require.NotNil(t, r.Update)
}

func TestValidateGatewayConfiguration(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty", "", false},
		{"name", "my-config", false},
		{"versioned", "my-config:2", true},
		{"latest", "my-config:latest", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := validateGatewayConfiguration(tc.value, "configuration")
			if tc.wantErr {
				require.Len(t, errs, 1)
				require.Contains(t, errs[0].Error(), "without a version")
				return
			}
			require.Empty(t, errs)
		})
	}
}

func TestResourceGatewayCreate(t *testing.T) {
	f := newFakeBindplane(t)
	f.put(testConfigurationResource("gw-config", map[string]string{"role": "gateway"}))

	d := newGatewayData(t, map[string]any{
		"name":                   "my-gateway",
		"display_name":           "My Gateway",
		"description":            "VMs behind a load balancer",
		"configuration":          "gw-config",
		"load_balancer_hostname": "lb.example.com",
	})
	require.NoError(t, resourceGatewayCreate(d, f.client()))
	require.Equal(t, "my-gateway", d.Id())

	applied := f.lastApplied()
	require.Equal(t, "Gateway", applied["kind"])
	require.Equal(t, "bindplane.observiq.com/v1", applied["apiVersion"])
	require.NotContains(t, applied, "status")

	metadata := applied["metadata"].(map[string]any)
	require.Equal(t, "my-gateway", metadata["id"])
	require.Equal(t, "my-gateway", metadata["name"])
	require.Equal(t, "My Gateway", metadata["displayName"])
	require.Equal(t, "VMs behind a load balancer", metadata["description"])

	spec := applied["spec"].(map[string]any)
	require.Equal(t, map[string]any{
		"configuration":        "gw-config",
		"loadBalancerHostname": "lb.example.com",
	}, spec)
	require.NotContains(t, spec, "workload")

	// State is read back from Bindplane
	require.Equal(t, "My Gateway", d.Get("display_name"))
	require.Equal(t, "VMs behind a load balancer", d.Get("description"))
	require.Equal(t, "gw-config", d.Get("configuration"))
	require.Equal(t, "lb.example.com", d.Get("load_balancer_hostname"))
}

func TestResourceGatewayCreateMinimal(t *testing.T) {
	f := newFakeBindplane(t)

	d := newGatewayData(t, map[string]any{"name": "my-gateway"})
	require.NoError(t, resourceGatewayCreate(d, f.client()))
	require.Equal(t, "my-gateway", d.Id())

	applied := f.lastApplied()
	require.Equal(t, map[string]any{
		"configuration":        "",
		"loadBalancerHostname": "",
	}, applied["spec"])

	require.Equal(t, "", d.Get("display_name"))
	require.Equal(t, "", d.Get("configuration"))
}

func TestResourceGatewayCreateErrors(t *testing.T) {
	t.Run("gateway exists", func(t *testing.T) {
		f := newFakeBindplane(t)
		f.put(testGatewayResource("01ABC", "my-gateway", map[string]any{}))

		d := newGatewayData(t, map[string]any{"name": "my-gateway"})
		err := resourceGatewayCreate(d, f.client())
		require.ErrorContains(t, err, "gateway with name 'my-gateway' already exists")
		require.Equal(t, "", d.Id())
	})

	t.Run("gateway lookup fails", func(t *testing.T) {
		f := newFakeBindplane(t)
		f.fail("/v1/resources/Gateway/my-gateway", http.StatusInternalServerError)

		d := newGatewayData(t, map[string]any{"name": "my-gateway"})
		require.ErrorContains(t, resourceGatewayCreate(d, f.client()), "failed to get gateway with name my-gateway")
	})

	t.Run("configuration does not exist", func(t *testing.T) {
		f := newFakeBindplane(t)

		d := newGatewayData(t, map[string]any{"name": "my-gateway", "configuration": "missing"})
		err := resourceGatewayCreate(d, f.client())
		require.ErrorContains(t, err, "configuration 'missing' does not exist")
		require.Nil(t, f.get(model.KindGateway, "my-gateway"))
	})

	t.Run("configuration lookup fails", func(t *testing.T) {
		f := newFakeBindplane(t)
		f.fail("/v1/configurations/broken", http.StatusInternalServerError)

		d := newGatewayData(t, map[string]any{"name": "my-gateway", "configuration": "broken"})
		require.ErrorContains(t, resourceGatewayCreate(d, f.client()), "failed to get configuration with name broken")
	})

	t.Run("apply fails", func(t *testing.T) {
		f := newFakeBindplane(t)
		f.applyStatus = model.UpdateStatus("invalid")

		d := newGatewayData(t, map[string]any{"name": "my-gateway"})
		err := resourceGatewayCreate(d, f.client())
		require.ErrorContains(t, err, "status: invalid")
		require.NotContains(t, err.Error(), "v1.103.0")
	})

	t.Run("bindplane too old for lookup", func(t *testing.T) {
		// Response from Bindplane v1.96.0
		f := newFakeBindplane(t)
		f.failWithBody("/v1/resources/Gateway/my-gateway", http.StatusBadRequest,
			`{"errors":["kind: Gateway cannot be requested using the resource endpoint"]}`)

		d := newGatewayData(t, map[string]any{"name": "my-gateway"})
		err := resourceGatewayCreate(d, f.client())
		require.ErrorContains(t, err, "gateways require Bindplane v1.103.0 or newer")
	})

	t.Run("bindplane too old for apply", func(t *testing.T) {
		// Response from Bindplane v1.96.0
		f := newFakeBindplane(t)
		f.failWithBody("/v1/apply", http.StatusBadRequest, `{"errors":["unknown resource kind: Gateway"]}`)

		d := newGatewayData(t, map[string]any{"name": "my-gateway"})
		err := resourceGatewayCreate(d, f.client())
		require.ErrorContains(t, err, "gateways require Bindplane v1.103.0 or newer")
		require.ErrorContains(t, err, "unknown resource kind: Gateway")
	})
}

func TestResourceGatewayUpdate(t *testing.T) {
	f := newFakeBindplane(t)
	f.put(testConfigurationResource("config-a", nil))
	f.put(testConfigurationResource("config-b", nil))

	d := newGatewayData(t, map[string]any{
		"name":                   "my-gateway",
		"display_name":           "Before",
		"description":            "before",
		"configuration":          "config-a",
		"load_balancer_hostname": "a.example.com",
	})
	require.NoError(t, resourceGatewayCreate(d, f.client()))

	// Update every attribute other than name, in place
	require.NoError(t, d.Set("display_name", "After"))
	require.NoError(t, d.Set("description", ""))
	require.NoError(t, d.Set("configuration", "config-b"))
	require.NoError(t, d.Set("load_balancer_hostname", "b.example.com"))
	require.NoError(t, resourceGatewayCreate(d, f.client()))
	require.Equal(t, "my-gateway", d.Id())

	applied := f.lastApplied()
	metadata := applied["metadata"].(map[string]any)
	require.Equal(t, "my-gateway", metadata["id"])
	require.Equal(t, "After", metadata["displayName"])
	require.NotContains(t, metadata, "description")
	require.Equal(t, map[string]any{
		"configuration":        "config-b",
		"loadBalancerHostname": "b.example.com",
	}, applied["spec"])

	require.Equal(t, "After", d.Get("display_name"))
	require.Equal(t, "", d.Get("description"))
	require.Equal(t, "config-b", d.Get("configuration"))
	require.Equal(t, "b.example.com", d.Get("load_balancer_hostname"))
}

func TestResourceGatewayUpdateKeepsBindplaneID(t *testing.T) {
	// A gateway created outside of Terraform has a Bindplane assigned ID.
	// After import, updates must keep that ID.
	f := newFakeBindplane(t)
	f.put(testGatewayResource("01BINDPLANEID", "my-gateway", map[string]any{}))

	d := newGatewayData(t, map[string]any{"name": "my-gateway", "display_name": "Imported"})
	d.SetId("01BINDPLANEID")
	require.NoError(t, resourceGatewayCreate(d, f.client()))

	metadata := f.lastApplied()["metadata"].(map[string]any)
	require.Equal(t, "01BINDPLANEID", metadata["id"])
	require.Equal(t, "01BINDPLANEID", d.Id())
	require.Equal(t, "Imported", d.Get("display_name"))
}

func TestResourceGatewayRead(t *testing.T) {
	t.Run("sets every field", func(t *testing.T) {
		f := newFakeBindplane(t)
		r := testGatewayResource("my-gateway", "my-gateway", map[string]any{
			"configuration":        "gw-config",
			"loadBalancerHostname": "lb.example.com",
		})
		r.Metadata.DisplayName = "Display"
		r.Metadata.Description = "Description"
		f.put(r)

		d := newGatewayData(t, map[string]any{"name": "my-gateway"})
		d.SetId("my-gateway")
		require.NoError(t, resourceGatewayRead(d, f.client()))
		require.Equal(t, "my-gateway", d.Id())
		require.Equal(t, "my-gateway", d.Get("name"))
		require.Equal(t, "Display", d.Get("display_name"))
		require.Equal(t, "Description", d.Get("description"))
		require.Equal(t, "gw-config", d.Get("configuration"))
		require.Equal(t, "lb.example.com", d.Get("load_balancer_hostname"))
	})

	t.Run("drift to empty values", func(t *testing.T) {
		// Values cleared in Bindplane must be cleared in state
		f := newFakeBindplane(t)
		f.put(testGatewayResource("my-gateway", "my-gateway", map[string]any{}))

		d := newGatewayData(t, map[string]any{
			"name":                   "my-gateway",
			"display_name":           "Display",
			"description":            "Description",
			"configuration":          "gw-config",
			"load_balancer_hostname": "lb.example.com",
		})
		d.SetId("my-gateway")
		require.NoError(t, resourceGatewayRead(d, f.client()))
		require.Equal(t, "", d.Get("display_name"))
		require.Equal(t, "", d.Get("description"))
		require.Equal(t, "", d.Get("configuration"))
		require.Equal(t, "", d.Get("load_balancer_hostname"))
	})

	t.Run("not found clears id", func(t *testing.T) {
		f := newFakeBindplane(t)

		d := newGatewayData(t, map[string]any{"name": "my-gateway"})
		d.SetId("my-gateway")
		require.NoError(t, resourceGatewayRead(d, f.client()))
		require.Equal(t, "", d.Id())
	})

	t.Run("different id clears id", func(t *testing.T) {
		f := newFakeBindplane(t)
		f.put(testGatewayResource("other-id", "my-gateway", map[string]any{}))

		d := newGatewayData(t, map[string]any{"name": "my-gateway"})
		d.SetId("my-gateway")
		require.NoError(t, resourceGatewayRead(d, f.client()))
		require.Equal(t, "", d.Id())
	})

	t.Run("server error", func(t *testing.T) {
		f := newFakeBindplane(t)
		f.fail("/v1/resources/Gateway/my-gateway", http.StatusInternalServerError)

		d := newGatewayData(t, map[string]any{"name": "my-gateway"})
		d.SetId("my-gateway")
		require.Error(t, resourceGatewayRead(d, f.client()))
		require.Equal(t, "my-gateway", d.Id())
	})

	t.Run("operator mode", func(t *testing.T) {
		f := newFakeBindplane(t)
		f.put(testGatewayResource("my-gateway", "my-gateway", map[string]any{
			"workload": map[string]any{"operator": "op", "workload": "wl"},
		}))

		d := newGatewayData(t, map[string]any{"name": "my-gateway"})
		d.SetId("my-gateway")
		require.ErrorContains(t, resourceGatewayRead(d, f.client()), "managed by the Bindplane Kubernetes operator")
	})
}

func TestResourceGatewayDelete(t *testing.T) {
	f := newFakeBindplane(t)
	f.put(testGatewayResource("my-gateway", "my-gateway", map[string]any{}))

	d := newGatewayData(t, map[string]any{"name": "my-gateway"})
	d.SetId("my-gateway")
	require.NoError(t, resourceGatewayDelete(d, f.client()))
	require.Equal(t, "", d.Id())
	require.Nil(t, f.get(model.KindGateway, "my-gateway"))

	// Deleting a missing gateway returns the error
	d.SetId("my-gateway")
	require.ErrorContains(t, resourceGatewayDelete(d, f.client()), "error while deleting gateway with name my-gateway")
	require.Equal(t, "my-gateway", d.Id())
}

func TestResourceGatewayImport(t *testing.T) {
	t.Run("created by terraform", func(t *testing.T) {
		f := newFakeBindplane(t)
		f.put(testGatewayResource("my-gateway", "my-gateway", map[string]any{"configuration": "gw-config"}))

		d := newGatewayData(t, map[string]any{})
		d.SetId("my-gateway")
		out, err := resourceGatewayImportState(context.Background(), d, f.client())
		require.NoError(t, err)
		require.Len(t, out, 1)
		require.Equal(t, "my-gateway", out[0].Id())
		require.Equal(t, "my-gateway", out[0].Get("name"))

		require.NoError(t, resourceGatewayRead(out[0], f.client()))
		require.Equal(t, "my-gateway", out[0].Id())
		require.Equal(t, "gw-config", out[0].Get("configuration"))
	})

	t.Run("created outside terraform", func(t *testing.T) {
		f := newFakeBindplane(t)
		r := testGatewayResource("01BINDPLANEID", "my-gateway", map[string]any{"loadBalancerHostname": "lb"})
		r.Metadata.DisplayName = "UI Gateway"
		f.put(r)

		d := newGatewayData(t, map[string]any{})
		d.SetId("my-gateway")
		out, err := resourceGatewayImportState(context.Background(), d, f.client())
		require.NoError(t, err)
		require.Equal(t, "01BINDPLANEID", out[0].Id())

		require.NoError(t, resourceGatewayRead(out[0], f.client()))
		require.Equal(t, "01BINDPLANEID", out[0].Id())
		require.Equal(t, "UI Gateway", out[0].Get("display_name"))
		require.Equal(t, "lb", out[0].Get("load_balancer_hostname"))
	})

	t.Run("not found", func(t *testing.T) {
		f := newFakeBindplane(t)

		d := newGatewayData(t, map[string]any{})
		d.SetId("missing")
		_, err := resourceGatewayImportState(context.Background(), d, f.client())
		require.ErrorContains(t, err, "gateway with name 'missing' does not exist")
	})

	t.Run("server error", func(t *testing.T) {
		f := newFakeBindplane(t)
		f.fail("/v1/resources/Gateway/my-gateway", http.StatusInternalServerError)

		d := newGatewayData(t, map[string]any{})
		d.SetId("my-gateway")
		_, err := resourceGatewayImportState(context.Background(), d, f.client())
		require.Error(t, err)
	})

	t.Run("operator mode", func(t *testing.T) {
		f := newFakeBindplane(t)
		f.put(testGatewayResource("01ABC", "my-gateway", map[string]any{
			"workload": map[string]any{"operator": "op", "workload": "wl"},
		}))

		d := newGatewayData(t, map[string]any{})
		d.SetId("my-gateway")
		_, err := resourceGatewayImportState(context.Background(), d, f.client())
		require.ErrorContains(t, err, "managed by the Bindplane Kubernetes operator")
	})
}
