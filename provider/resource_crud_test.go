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
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/terraform-provider-bindplane/internal/fakebindplane"
	"github.com/stretchr/testify/require"
)

// seedConfiguration is a configuration shaped like the Bindplane server's
// responses, with versioned resource names and fields the provider does not use.
const seedConfiguration = `{
	"apiVersion": "bindplane.observiq.com/v2",
	"kind": "Configuration",
	"metadata": {
		"id": "cfg-id",
		"name": "my-config",
		"version": 2,
		"hash": "abc123",
		"labels": {"platform": "linux", "env": "test"}
	},
	"spec": {
		"contentType": "text/yaml",
		"measurementInterval": "1m",
		"parameters": [
			{"name": "telemetryPort", "value": 8888},
			{"name": "telemetryLevel", "value": "detailed"}
		],
		"sources": [{
			"id": "src-route",
			"name": "my-source:3",
			"processors": [{"name": "my-batch:1"}],
			"routes": {"logs": [{"id": "r1", "components": ["destinations/dest-route"]}]}
		}],
		"processors": [{
			"id": "pg-1",
			"processors": [{"name": "my-batch:1"}, {"name": "my-filter:4"}],
			"routes": {"metrics": [{"id": "r2", "components": ["destinations/dest-route"]}]}
		}],
		"connectors": [{
			"id": "conn-1",
			"name": "my-connector:1",
			"routes": {"traces": [{"id": "r3", "components": ["processors/pg-1"]}]}
		}],
		"destinations": [{"id": "dest-route", "name": "my-dest:2", "processors": [{"name": "my-filter:1"}]}],
		"extensions": [{"name": "my-pprof:1"}],
		"selector": {"matchLabels": {"configuration": "my-config"}},
		"rollout": {
			"type": "progressive",
			"parameters": [{"name": "stages", "value": [{"labels": {"env": "stage"}, "name": "stage"}]}]
		}
	},
	"status": {"rollout": {"status": 4}}
}`

func TestComponentResources(t *testing.T) {
	cases := []struct {
		kind     string
		resource func() *schema.Resource
		path     string
	}{
		{"Source", resourceSource, "/v1/sources/my-component"},
		{"Destination", resourceDestination, "/v1/destinations/my-component"},
		{"Processor", resourceProcessor, "/v1/processors/my-component"},
		{"Extension", resourceExtension, "/v1/extensions/my-component"},
		{"Connector", resourceConnector, "/v1/resources/Connector/my-component"},
	}

	const parametersJSON = `[{"name":"telemetry_types","value":["Logs","Metrics"]},{"name":"interval","value":30},{"name":"enabled","value":true}]`

	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			srv := fakebindplane.New(t)
			meta := newProviderMeta(t, srv, nil)
			r := tc.resource()

			// Create
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
				"name":            "my-component",
				"type":            "custom",
				"parameters_json": parametersJSON,
				"rollout":         true,
			})
			require.NoError(t, r.Create(d, meta))

			require.Regexp(t, "^"+generatedIDPattern.String()+"$", d.Id())
			require.Equal(t, d.Id(), srv.Resource(tc.kind, "my-component")["metadata"].(map[string]any)["id"])
			require.JSONEq(t, `{"resources": [{
				"apiVersion": "bindplane.observiq.com/v1",
				"kind": "`+tc.kind+`",
				"metadata": {"id": "<generated>", "name": "my-component", "labels": {}},
				"spec": {
					"type": "custom",
					"parameters": [
						{"name": "telemetry_types", "value": ["Logs", "Metrics"]},
						{"name": "interval", "value": 30},
						{"name": "enabled", "value": true}
					],
					"processors": []
				}
			}]}`, lastApplyBody(t, srv))
			for _, req := range srv.Requests() {
				require.False(t, strings.HasPrefix(req.Path, "/v1/rollouts/"), "only configurations are rolled out")
			}

			// Create reads the resource back into state
			require.NotNil(t, srv.LastRequest(http.MethodGet, tc.path))
			require.Equal(t, "my-component", d.Get("name"))
			require.Equal(t, "custom", d.Get("type"))
			require.JSONEq(t, parametersJSON, d.Get("parameters_json").(string))

			// Update re-applies with the existing ID
			id := d.Id()
			require.NoError(t, d.Set("parameters_json", `[{"name":"interval","value":60}]`))
			require.NoError(t, r.Update(d, meta))
			require.Equal(t, id, d.Id())
			require.JSONEq(t, `{"resources": [{
				"apiVersion": "bindplane.observiq.com/v1",
				"kind": "`+tc.kind+`",
				"metadata": {"id": "<generated>", "name": "my-component", "labels": {}},
				"spec": {"type": "custom", "parameters": [{"name": "interval", "value": 60}], "processors": []}
			}]}`, lastApplyBody(t, srv))
			require.JSONEq(t, `[{"name":"interval","value":60}]`, d.Get("parameters_json").(string))

			// Creating a resource that already exists is an error
			existing := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
				"name":    "my-component",
				"type":    "custom",
				"rollout": false,
			})
			require.ErrorContains(t, r.Create(existing, meta), "already exists")

			// Import by name
			imported := r.TestResourceData()
			imported.SetId("my-component")
			states, err := r.Importer.StateContext(context.Background(), imported, meta)
			require.NoError(t, err)
			require.Len(t, states, 1)
			require.Equal(t, id, states[0].Id())
			require.Equal(t, "my-component", states[0].Get("name"))

			missing := r.TestResourceData()
			missing.SetId("missing")
			_, err = r.Importer.StateContext(context.Background(), missing, meta)
			require.ErrorContains(t, err, "does not exist")

			// Delete
			require.NoError(t, r.Delete(d, meta))
			req := srv.LastRequest(http.MethodDelete, tc.path)
			require.NotNil(t, req)
			require.Equal(t, "force=false", req.Query.Encode())
			require.Nil(t, srv.Resource(tc.kind, "my-component"))

			// Reading a deleted resource removes it from state
			require.NoError(t, r.Read(d, meta))
			require.Empty(t, d.Id())
		})
	}
}

func TestGenericResourceRead(t *testing.T) {
	const seedSource = `{
		"apiVersion": "bindplane.observiq.com/v1",
		"kind": "Source",
		"metadata": {"id": "src-id", "name": "my-source", "version": 3, "labels": {}},
		"spec": {
			"type": "custom:2",
			"parameters": [
				{"name": "password", "value": "(sensitive)", "sensitive": true},
				{"name": "token", "value": "(sensitive)", "sensitive": true},
				{"name": "host", "value": "localhost"},
				{"name": "api_key", "value": "abc", "credential": true}
			]
		}
	}`

	t.Run("sensitive values are kept from state", func(t *testing.T) {
		srv := fakebindplane.New(t)
		meta := newProviderMeta(t, srv, nil)
		srv.Seed(t, seedSource)

		d := schema.TestResourceDataRaw(t, resourceSource().Schema, map[string]any{
			"name":            "my-source",
			"type":            "custom",
			"rollout":         false,
			"parameters_json": `[{"name":"password","value":"hunter2","sensitive":true},{"name":"token","value":"xyz"},{"name":"host","value":"old"}]`,
		})
		d.SetId("src-id")

		require.NoError(t, resourceSourceRead(d, meta))
		require.Equal(t, "src-id", d.Id())
		require.Equal(t, "my-source", d.Get("name"))
		require.Equal(t, "custom", d.Get("type"), "the type version should be removed")
		require.JSONEq(t, `[
			{"name": "password", "value": "hunter2", "sensitive": true},
			{"name": "token", "value": "xyz"},
			{"name": "host", "value": "localhost"}
		]`, d.Get("parameters_json").(string), "parameters not in state should not be saved")
	})

	t.Run("id mismatch", func(t *testing.T) {
		srv := fakebindplane.New(t)
		meta := newProviderMeta(t, srv, nil)
		srv.Seed(t, seedSource)

		d := schema.TestResourceDataRaw(t, resourceSource().Schema, map[string]any{
			"name":    "my-source",
			"type":    "custom",
			"rollout": false,
		})
		d.SetId("created-elsewhere")

		require.NoError(t, resourceSourceRead(d, meta))
		require.Empty(t, d.Id(), "a resource with a different ID should be re-created")
	})

	t.Run("server error", func(t *testing.T) {
		srv := fakebindplane.New(t)
		meta := newProviderMeta(t, srv, nil)
		srv.RespondOnce(http.MethodGet, "/v1/sources/my-source", fakebindplane.Response{
			Status: http.StatusInternalServerError,
			Body:   `{"errors": ["boom"]}`,
		})

		d := schema.TestResourceDataRaw(t, resourceSource().Schema, map[string]any{
			"name":    "my-source",
			"type":    "custom",
			"rollout": false,
		})
		d.SetId("src-id")

		require.ErrorContains(t, resourceSourceRead(d, meta), "boom")
	})
}

func TestResourceProcessorBundle(t *testing.T) {
	srv := fakebindplane.New(t)
	meta := newProviderMeta(t, srv, nil)
	r := resourceProcessorBundle()

	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"name":    "my-bundle",
		"rollout": false,
		"processor": []any{
			map[string]any{"name": "my-batch"},
			map[string]any{"name": "my-filter"},
		},
	})
	require.NoError(t, r.Create(d, meta))

	require.Regexp(t, "^"+generatedIDPattern.String()+"$", d.Id())
	require.JSONEq(t, `{"resources": [{
		"apiVersion": "bindplane.observiq.com/v1",
		"kind": "Processor",
		"metadata": {"id": "<generated>", "name": "my-bundle", "labels": {}},
		"spec": {
			"type": "processor_bundle",
			"parameters": null,
			"processors": [{"name": "my-batch"}, {"name": "my-filter"}]
		}
	}]}`, lastApplyBody(t, srv))

	require.Equal(t, "my-bundle", d.Get("name"))
	require.Equal(t, "processor_bundle", d.Get("type"))
	require.Equal(t, []any{
		map[string]any{"name": "my-batch"},
		map[string]any{"name": "my-filter"},
	}, d.Get("processor"))

	// Read strips resource versions from processor names
	srv.Seed(t, `{
		"apiVersion": "bindplane.observiq.com/v1",
		"kind": "Processor",
		"metadata": {"id": "`+d.Id()+`", "name": "my-bundle", "labels": {}},
		"spec": {"type": "processor_bundle:1", "processors": [{"name": "my-batch:2"}, {"name": "my-filter:7"}]}
	}`)
	require.NoError(t, r.Read(d, meta))
	require.Equal(t, "processor_bundle", d.Get("type"))
	require.Equal(t, []any{
		map[string]any{"name": "my-batch"},
		map[string]any{"name": "my-filter"},
	}, d.Get("processor"))

	imported := r.TestResourceData()
	imported.SetId("my-bundle")
	states, err := r.Importer.StateContext(context.Background(), imported, meta)
	require.NoError(t, err)
	require.Equal(t, d.Id(), states[0].Id())

	require.NoError(t, r.Delete(d, meta))
	require.NotNil(t, srv.LastRequest(http.MethodDelete, "/v1/processors/my-bundle"))
}

func TestResourceFleet(t *testing.T) {
	srv := fakebindplane.New(t)
	meta := newProviderMeta(t, srv, nil)
	r := resourceFleet()

	raw := map[string]any{
		"name":          "my-fleet",
		"display_name":  "My Fleet",
		"agent_type":    "observiq-otel-collector",
		"platform":      "linux",
		"configuration": "my-config",
	}

	// The configuration must exist
	d := schema.TestResourceDataRaw(t, r.Schema, raw)
	require.ErrorContains(t, r.Create(d, meta), "configuration 'my-config' does not exist")
	require.Empty(t, srv.RequestsTo(http.MethodPost, "/v1/apply"))

	srv.Seed(t, seedConfiguration)

	d = schema.TestResourceDataRaw(t, r.Schema, raw)
	require.NoError(t, r.Create(d, meta))
	require.Equal(t, "my-fleet", d.Id())
	require.JSONEq(t, `{"resources": [{
		"apiVersion": "bindplane.observiq.com/v1",
		"kind": "Fleet",
		"metadata": {
			"id": "my-fleet",
			"name": "my-fleet",
			"displayName": "My Fleet",
			"labels": {"agent-type": "observiq-otel-collector", "platform": "linux"}
		},
		"spec": {
			"configuration": "my-config",
			"selector": {"matchLabels": {"fleet": "my-fleet"}}
		}
	}]}`, lastApplyBody(t, srv))
	require.NotNil(t, srv.LastRequest(http.MethodGet, "/v1/resources/Fleet/my-fleet"))

	// Creating a fleet that already exists is an error
	existing := schema.TestResourceDataRaw(t, r.Schema, raw)
	require.ErrorContains(t, r.Create(existing, meta), "already exists")

	// Read picks up changes made outside of Terraform
	srv.Seed(t, `{
		"apiVersion": "bindplane.observiq.com/v1",
		"kind": "Fleet",
		"metadata": {
			"id": "my-fleet",
			"name": "my-fleet",
			"displayName": "Renamed",
			"labels": {"agent-type": "bdot", "platform": "windows", "other": "label"}
		},
		"spec": {"configuration": "other-config", "selector": {"matchLabels": {"fleet": "my-fleet"}}}
	}`)
	require.NoError(t, r.Read(d, meta))
	require.Equal(t, "my-fleet", d.Id())
	require.Equal(t, "Renamed", d.Get("display_name"))
	require.Equal(t, "bdot", d.Get("agent_type"))
	require.Equal(t, "windows", d.Get("platform"))
	require.Equal(t, "other-config", d.Get("configuration"))

	imported := r.TestResourceData()
	imported.SetId("my-fleet")
	states, err := r.Importer.StateContext(context.Background(), imported, meta)
	require.NoError(t, err)
	require.Equal(t, "my-fleet", states[0].Id())
	require.Equal(t, "my-fleet", states[0].Get("name"))

	require.NoError(t, r.Delete(d, meta))
	req := srv.LastRequest(http.MethodDelete, "/v1/resources/Fleet/my-fleet")
	require.NotNil(t, req)
	require.Equal(t, "force=false", req.Query.Encode())
	require.Empty(t, d.Id())

	require.NoError(t, r.Read(d, meta))
	require.Empty(t, d.Id())
}

func configurationTestRaw() map[string]any {
	return map[string]any{
		"name":                 "my-config",
		"platform":             "linux",
		"labels":               map[string]any{"env": "test"},
		"measurement_interval": "1m",
		"rollout":              true,
		"rollout_options": []any{map[string]any{
			"type": "progressive",
			"parameters": []any{map[string]any{
				"name": "stages",
				"value": []any{
					map[string]any{"name": "stage", "labels": map[string]any{"env": "stage"}},
					map[string]any{"name": "production", "labels": map[string]any{"env": "production"}},
				},
			}},
		}},
		"advanced": []any{map[string]any{
			"metrics": []any{map[string]any{"port": 8888, "level": "detailed"}},
		}},
		"extensions": []any{"my-pprof"},
	}
}

func TestResourceConfiguration(t *testing.T) {
	srv := fakebindplane.New(t)
	meta := newProviderMeta(t, srv, nil)
	r := resourceConfiguration()

	raw := configurationTestRaw()
	raw["source"] = []any{map[string]any{"name": "my-source", "processors": []any{"my-batch", "my-filter"}}}
	raw["destination"] = []any{map[string]any{"name": "my-dest", "processors": []any{"my-filter"}}}

	d := schema.TestResourceDataRaw(t, r.Schema, raw)
	require.NoError(t, r.Create(d, meta))

	requireGoldenJSON(t, "configuration_v1_apply.json", lastApplyBody(t, srv))

	rollout := srv.LastRequest(http.MethodPost, "/v1/rollouts/my-config/start")
	require.NotNil(t, rollout)
	require.JSONEq(t, `{"options": null}`, string(rollout.Body))

	// The ID is assigned by Bindplane
	require.Equal(t, srv.Resource("Configuration", "my-config")["metadata"].(map[string]any)["id"], d.Id())
	require.NotEmpty(t, d.Id())

	// Create reads the configuration back into state
	require.Equal(t, "linux", d.Get("platform"))
	require.Equal(t, map[string]any{"env": "test"}, d.Get("labels"))
	require.Equal(t, map[string]any{"configuration": "my-config"}, d.Get("match_labels"))
	require.Equal(t, []any{map[string]any{"name": "my-source", "processors": []any{"my-batch", "my-filter"}}}, d.Get("source"))
	require.Equal(t, []any{map[string]any{"name": "my-dest", "processors": []any{"my-filter"}}}, d.Get("destination"))
	require.Equal(t, []any{"my-pprof"}, d.Get("extensions"))
	require.Equal(t, "1m", d.Get("measurement_interval"))
	require.Equal(t, 8888, d.Get("advanced.0.metrics.0.port"))
	require.Equal(t, "detailed", d.Get("advanced.0.metrics.0.level"))
	require.Equal(t, "progressive", d.Get("rollout_options.0.type"))
	require.Equal(t, "stages", d.Get("rollout_options.0.parameters.0.name"))
	require.Equal(t, "production", d.Get("rollout_options.0.parameters.0.value.1.name"))

	// Creating a configuration that already exists is an error
	existing := schema.TestResourceDataRaw(t, r.Schema, raw)
	require.ErrorContains(t, r.Create(existing, meta), "already exists")

	// Delete
	require.NoError(t, r.Delete(d, meta))
	require.NotNil(t, srv.LastRequest(http.MethodDelete, "/v1/configurations/my-config"))
	require.NoError(t, r.Read(d, meta))
	require.Empty(t, d.Id())
}

func TestResourceConfiguration_read(t *testing.T) {
	srv := fakebindplane.New(t)
	meta := newProviderMeta(t, srv, nil)
	srv.Seed(t, seedConfiguration)
	r := resourceConfiguration()

	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"name":     "my-config",
		"platform": "linux",
		"rollout":  false,
	})
	d.SetId("cfg-id")

	require.NoError(t, r.Read(d, meta))
	require.Equal(t, "cfg-id", d.Id())
	require.Equal(t, "my-config", d.Get("name"))
	require.Equal(t, "linux", d.Get("platform"))
	require.Equal(t, map[string]any{"env": "test"}, d.Get("labels"), "the platform label is stored separately")
	require.Equal(t, map[string]any{"configuration": "my-config"}, d.Get("match_labels"))
	require.Equal(t, []any{map[string]any{"name": "my-source", "processors": []any{"my-batch"}}}, d.Get("source"))
	require.Equal(t, []any{map[string]any{"name": "my-dest", "processors": []any{"my-filter"}}}, d.Get("destination"))
	require.Equal(t, []any{"my-pprof"}, d.Get("extensions"))
	require.Equal(t, "1m", d.Get("measurement_interval"))
	require.Equal(t, 8888, d.Get("advanced.0.metrics.0.port"))
	require.Equal(t, "detailed", d.Get("advanced.0.metrics.0.level"))
	require.Equal(t, "progressive", d.Get("rollout_options.0.type"))
	require.Equal(t, "stage", d.Get("rollout_options.0.parameters.0.value.0.name"))
	require.Equal(t, map[string]any{"env": "stage"}, d.Get("rollout_options.0.parameters.0.value.0.labels"))

	// A configuration with a different ID was not created by Terraform
	d.SetId("other-id")
	require.NoError(t, r.Read(d, meta))
	require.Empty(t, d.Id())
}

// configurationV2TestRaw returns a bindplane_configuration_v2 config
// with every component type and their routes.
func configurationV2TestRaw() map[string]any {
	raw := configurationTestRaw()
	raw["source"] = []any{map[string]any{
		"name":       "my-source",
		"processors": []any{"my-batch"},
		"route": []any{
			map[string]any{"route_id": "source-logs", "telemetry_type": "logs", "components": []any{"processors/pg-1"}},
			map[string]any{"route_id": "source-traces", "telemetry_type": "traces", "components": []any{"connectors/conn-1"}},
		},
	}}
	raw["processor_group"] = []any{map[string]any{
		"route_id":   "pg-1",
		"processors": []any{"my-filter", "my-batch"},
		"route": []any{
			map[string]any{"route_id": "pg-logs", "telemetry_type": "logs", "components": []any{"destinations/dest-1"}},
		},
	}}
	raw["connector"] = []any{map[string]any{
		"route_id": "conn-1",
		"name":     "my-connector",
		"route": []any{
			map[string]any{"route_id": "conn-metrics", "telemetry_type": "metrics", "components": []any{"destinations/dest-1"}},
		},
	}}
	raw["destination"] = []any{map[string]any{
		"route_id":   "dest-1",
		"name":       "my-dest",
		"processors": []any{"my-filter"},
	}}
	return raw
}

func TestResourceConfigurationV2(t *testing.T) {
	srv := fakebindplane.New(t)
	meta := newProviderMeta(t, srv, nil)
	r := resourceConfigurationV2()

	d := schema.TestResourceDataRaw(t, r.Schema, configurationV2TestRaw())
	require.NoError(t, r.Create(d, meta))

	requireGoldenJSON(t, "configuration_v2_apply.json", lastApplyBody(t, srv))
	require.NotNil(t, srv.LastRequest(http.MethodPost, "/v1/rollouts/my-config/start"))
	require.NotEmpty(t, d.Id())

	// Create reads the configuration back into state
	require.Equal(t, "my-source", d.Get("source.0.name"))
	require.Len(t, d.Get("source.0.route").(*schema.Set).List(), 2)
	require.Equal(t, "pg-1", d.Get("processor_group.0.route_id"))
	require.Equal(t, []any{"my-filter", "my-batch"}, d.Get("processor_group.0.processors"))
	require.Equal(t, "conn-1", d.Get("connector.0.route_id"))
	require.Equal(t, "my-connector", d.Get("connector.0.name"))
	require.Equal(t, "dest-1", d.Get("destination.0.route_id"))
	require.Equal(t, "my-dest", d.Get("destination.0.name"))
	require.Equal(t, []any{map[string]any{
		"route_id":       "conn-metrics",
		"telemetry_type": "metrics",
		"components":     []any{"destinations/dest-1"},
	}}, d.Get("connector.0.route").(*schema.Set).List())

	require.NoError(t, r.Delete(d, meta))
	require.NotNil(t, srv.LastRequest(http.MethodDelete, "/v1/configurations/my-config"))
}

func TestResourceConfigurationV2_read(t *testing.T) {
	srv := fakebindplane.New(t)
	meta := newProviderMeta(t, srv, nil)
	srv.Seed(t, seedConfiguration)
	r := resourceConfigurationV2()

	// Route IDs for processor groups and destinations are only
	// known to Terraform's state.
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"name":            "my-config",
		"platform":        "linux",
		"rollout":         false,
		"processor_group": []any{map[string]any{"route_id": "pg-1", "processors": []any{"my-batch"}}},
		"destination":     []any{map[string]any{"route_id": "dest-route", "name": "my-dest"}},
	})
	d.SetId("cfg-id")

	require.NoError(t, r.Read(d, meta))
	require.Equal(t, "cfg-id", d.Id())
	require.Equal(t, map[string]any{"env": "test"}, d.Get("labels"))

	require.Equal(t, "my-source", d.Get("source.0.name"))
	require.Equal(t, []any{"my-batch"}, d.Get("source.0.processors"))
	require.Equal(t, []any{map[string]any{
		"route_id":       "r1",
		"telemetry_type": "logs",
		"components":     []any{"destinations/dest-route"},
	}}, d.Get("source.0.route").(*schema.Set).List())

	require.Equal(t, "pg-1", d.Get("processor_group.0.route_id"))
	require.Equal(t, []any{"my-batch", "my-filter"}, d.Get("processor_group.0.processors"))
	require.Equal(t, []any{map[string]any{
		"route_id":       "r2",
		"telemetry_type": "metrics",
		"components":     []any{"destinations/dest-route"},
	}}, d.Get("processor_group.0.route").(*schema.Set).List())

	require.Equal(t, "conn-1", d.Get("connector.0.route_id"))
	require.Equal(t, "my-connector", d.Get("connector.0.name"))
	require.Equal(t, []any{map[string]any{
		"route_id":       "r3",
		"telemetry_type": "traces",
		"components":     []any{"processors/pg-1"},
	}}, d.Get("connector.0.route").(*schema.Set).List())

	require.Equal(t, "dest-route", d.Get("destination.0.route_id"))
	require.Equal(t, "my-dest", d.Get("destination.0.name"))
	require.Equal(t, []any{"my-filter"}, d.Get("destination.0.processors"))

	require.Equal(t, []any{"my-pprof"}, d.Get("extensions"))
	require.Equal(t, 8888, d.Get("advanced.0.metrics.0.port"))
	require.Equal(t, "progressive", d.Get("rollout_options.0.type"))
}

func TestResourceConfigurationV2_readRouteIDs(t *testing.T) {
	r := resourceConfigurationV2()

	t.Run("destination-state-differs", func(t *testing.T) {
		srv := fakebindplane.New(t)
		meta := newProviderMeta(t, srv, nil)
		srv.Seed(t, seedConfiguration)

		// A destination keeps the route ID saved in state, even when
		// Bindplane returns a different ID.
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
			"name":        "my-config",
			"destination": []any{map[string]any{"route_id": "state-route", "name": "my-dest"}},
		})
		require.NoError(t, r.Read(d, meta))
		require.Equal(t, "state-route", d.Get("destination.0.route_id"))
	})

	t.Run("from-bindplane", func(t *testing.T) {
		srv := fakebindplane.New(t)
		meta := newProviderMeta(t, srv, nil)
		srv.Seed(t, seedConfiguration)

		// Without prior state, route IDs come from Bindplane.
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"name": "my-config"})
		require.NoError(t, r.Read(d, meta))
		require.Equal(t, "pg-1", d.Get("processor_group.0.route_id"))
		require.Equal(t, "dest-route", d.Get("destination.0.route_id"))
	})

	t.Run("destination-state-fallback", func(t *testing.T) {
		srv := fakebindplane.New(t)
		meta := newProviderMeta(t, srv, nil)
		srv.Seed(t, strings.Replace(seedConfiguration, `"id": "dest-route", `, "", 1))

		// A destination without an ID keeps the route ID saved in state.
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
			"name":        "my-config",
			"destination": []any{map[string]any{"route_id": "state-route", "name": "my-dest"}},
		})
		require.NoError(t, r.Read(d, meta))
		require.Equal(t, "state-route", d.Get("destination.0.route_id"))
	})
}

func TestResourceConfigurationV2Import(t *testing.T) {
	srv := fakebindplane.New(t)
	meta := newProviderMeta(t, srv, nil)
	srv.Seed(t, seedConfiguration)
	r := resourceConfigurationV2()
	require.NotNil(t, r.Importer)

	// Terraform import sets the ID to the name given by the user.
	d := r.TestResourceData()
	d.SetId("my-config")

	imported, err := r.Importer.StateContext(context.Background(), d, meta)
	require.NoError(t, err)
	require.Len(t, imported, 1)
	require.Equal(t, "cfg-id", imported[0].Id())
	require.Equal(t, "my-config", imported[0].Get("name"))

	// Terraform reads the resource after import, with no prior state.
	d = imported[0]
	require.NoError(t, r.Read(d, meta))
	require.Equal(t, "cfg-id", d.Id())
	require.Equal(t, "my-config", d.Get("name"))
	require.Equal(t, "linux", d.Get("platform"))
	require.Equal(t, map[string]any{"env": "test"}, d.Get("labels"))
	require.Equal(t, map[string]any{"configuration": "my-config"}, d.Get("match_labels"))
	require.Equal(t, "", d.Get("description"))
	require.Equal(t, "1m", d.Get("measurement_interval"))

	require.Equal(t, "my-source", d.Get("source.0.name"))
	require.Equal(t, []any{"my-batch"}, d.Get("source.0.processors"))
	require.Len(t, d.Get("source.0.route").(*schema.Set).List(), 1)
	require.Equal(t, "pg-1", d.Get("processor_group.0.route_id"))
	require.Equal(t, []any{"my-batch", "my-filter"}, d.Get("processor_group.0.processors"))
	require.Equal(t, "conn-1", d.Get("connector.0.route_id"))
	require.Equal(t, "my-connector", d.Get("connector.0.name"))
	require.Equal(t, "dest-route", d.Get("destination.0.route_id"))
	require.Equal(t, "my-dest", d.Get("destination.0.name"))
	require.Equal(t, []any{"my-filter"}, d.Get("destination.0.processors"))
	require.Equal(t, []any{"my-pprof"}, d.Get("extensions"))
	require.Equal(t, 8888, d.Get("advanced.0.metrics.0.port"))
	require.Equal(t, "progressive", d.Get("rollout_options.0.type"))
	require.Equal(t, "stage", d.Get("rollout_options.0.parameters.0.value.0.name"))

	// Bindplane does not store rollout, so it is not imported.
	_, ok := d.GetOk("rollout")
	require.False(t, ok)

	t.Run("not-found", func(t *testing.T) {
		d := r.TestResourceData()
		d.SetId("missing")

		_, err := r.Importer.StateContext(context.Background(), d, meta)
		require.ErrorContains(t, err, "does not exist")
	})

	t.Run("server-error", func(t *testing.T) {
		srv.RespondOnce(http.MethodGet, "/v1/configurations/my-config", serverError)
		d := r.TestResourceData()
		d.SetId("my-config")

		_, err := r.Importer.StateContext(context.Background(), d, meta)
		require.ErrorContains(t, err, "boom")
	})
}

// TestResourceConfigurationV2ImportVerify checks that importing a
// configuration created by Terraform produces the same state,
// except for rollout which Bindplane does not store.
func TestResourceConfigurationV2ImportVerify(t *testing.T) {
	srv := fakebindplane.New(t)
	meta := newProviderMeta(t, srv, nil)
	r := resourceConfigurationV2()

	created := schema.TestResourceDataRaw(t, r.Schema, configurationV2TestRaw())
	require.NoError(t, r.Create(created, meta))

	d := r.TestResourceData()
	d.SetId("my-config")
	imported, err := r.Importer.StateContext(context.Background(), d, meta)
	require.NoError(t, err)
	require.Len(t, imported, 1)
	require.NoError(t, r.Read(imported[0], meta))

	want := created.State().Attributes
	got := imported[0].State().Attributes
	delete(want, "rollout")
	require.NotContains(t, got, "rollout")
	require.Equal(t, want, got)
}
