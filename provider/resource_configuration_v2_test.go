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
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/observiq/terraform-provider-bindplane/internal/fakebindplane"
	"github.com/observiq/terraform-provider-bindplane/internal/model"
	"github.com/stretchr/testify/require"
)

func TestProcessorsFromConfig(t *testing.T) {
	const batchParams = `[{"name":"send_batch_size","value":200}]`

	cases := []struct {
		name      string
		names     []any
		blocks    []any
		expect    []model.ResourceConfiguration
		expectErr string
	}{
		{
			name:   "none",
			expect: []model.ResourceConfiguration{},
		},
		{
			name:   "names",
			names:  []any{"my-filter", "my-batch"},
			expect: []model.ResourceConfiguration{{Name: "my-filter"}, {Name: "my-batch"}},
		},
		{
			name: "blocks",
			blocks: []any{
				map[string]any{"name": "my-filter", "display_name": "Filter", "id": "p-my-filter"},
				map[string]any{"type": "batch", "parameters_json": batchParams, "display_name": "Batch", "id": "batch-1"},
				map[string]any{"type": "batch", "id": ""},
			},
			expect: []model.ResourceConfiguration{
				// The ID of a processor referenced by name is not sent
				{Name: "my-filter", DisplayName: "Filter"},
				{
					ID:          "batch-1",
					DisplayName: "Batch",
					ParameterizedSpec: model.ParameterizedSpec{
						Type:       "batch",
						Parameters: []model.Parameter{{Name: "send_batch_size", Value: float64(200)}},
					},
				},
				{ParameterizedSpec: model.ParameterizedSpec{Type: "batch"}},
			},
		},
		{
			name:      "names and blocks",
			names:     []any{"my-filter"},
			blocks:    []any{map[string]any{"name": "my-batch"}},
			expectErr: "processors and processor cannot both be set",
		},
		{
			name:      "name and type",
			blocks:    []any{map[string]any{"name": "my-batch", "type": "batch"}},
			expectErr: "processor 0: name and type cannot both be set",
		},
		{
			name:      "empty block",
			blocks:    []any{map[string]any{"name": "my-batch"}, nil},
			expectErr: "processor 1: one of name or type must be set",
		},
		{
			name:      "parameters without type",
			blocks:    []any{map[string]any{"name": "my-batch", "parameters_json": batchParams}},
			expectErr: "processor 0: parameters_json requires type",
		},
		{
			name:      "invalid parameters",
			blocks:    []any{map[string]any{"type": "batch", "parameters_json": `{"name":"send_batch_size"}`}},
			expectErr: "failed to unmarshal parameters",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			processors, err := processorsFromConfig(tc.names, tc.blocks)
			if tc.expectErr != "" {
				require.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expect, processors)
		})
	}
}

func TestProcessorsToState(t *testing.T) {
	cases := []struct {
		name         string
		stateBlock   map[string]any
		processors   []model.ResourceConfiguration
		expectNames  []string
		expectBlocks []map[string]any
	}{
		{
			name:        "names",
			stateBlock:  map[string]any{"processors": []any{"my-filter"}},
			processors:  []model.ResourceConfiguration{{ID: "p-my-filter", Name: "my-filter:4"}},
			expectNames: []string{"my-filter"},
		},
		{
			name:        "no state",
			processors:  []model.ResourceConfiguration{{ID: "p-my-filter", Name: "my-filter:4"}},
			expectNames: []string{"my-filter"},
		},
		{
			name:        "no processors",
			stateBlock:  map[string]any{"processors": []any{"my-filter"}},
			expectNames: []string{},
		},
		{
			name:       "state uses blocks",
			stateBlock: map[string]any{"processor": []any{map[string]any{"name": "my-filter"}}},
			processors: []model.ResourceConfiguration{{ID: "p-my-filter", Name: "my-filter:4", DisplayName: "Filter"}},
			expectBlocks: []map[string]any{
				{"id": "p-my-filter", "name": "my-filter", "type": "", "parameters_json": "", "display_name": "Filter"},
			},
		},
		{
			// An inline processor added outside of Terraform can only be
			// represented by a processor block. Its parameters are not in
			// state, so none are kept, which shows as drift.
			name:       "inline processor not in state",
			stateBlock: map[string]any{"processors": []any{"my-filter"}},
			processors: []model.ResourceConfiguration{
				{ID: "p-my-filter", Name: "my-filter:4"},
				{
					ID: "batch-1",
					ParameterizedSpec: model.ParameterizedSpec{
						Type:       "batch:3",
						Parameters: []model.Parameter{{Name: "send_batch_size", Value: 200}},
					},
				},
			},
			expectBlocks: []map[string]any{
				{"id": "p-my-filter", "name": "my-filter", "type": "", "parameters_json": "", "display_name": ""},
				{"id": "batch-1", "name": "", "type": "batch", "parameters_json": "", "display_name": ""},
			},
		},
		{
			name: "parameters are reconciled with state",
			stateBlock: map[string]any{"processor": []any{
				map[string]any{"type": "batch", "parameters_json": `[{"name":"send_batch_size","value":200}]`},
				map[string]any{"type": "custom", "parameters_json": `[{"name":"api_key","value":"secret"}]`},
			}},
			processors: []model.ResourceConfiguration{
				{
					ID: "batch-1",
					ParameterizedSpec: model.ParameterizedSpec{
						Type: "batch:3",
						Parameters: []model.Parameter{
							{Name: "send_batch_size", Value: 200},
							{Name: "timeout", Value: "200ms"},
						},
					},
				},
				{
					ID: "custom-1",
					ParameterizedSpec: model.ParameterizedSpec{
						Type:       "custom:1",
						Parameters: []model.Parameter{{Name: "api_key", Value: "(sensitive)", Sensitive: true}},
					},
				},
			},
			expectBlocks: []map[string]any{
				{"id": "batch-1", "name": "", "type": "batch", "parameters_json": `[{"name":"send_batch_size","value":200}]`, "display_name": ""},
				{"id": "custom-1", "name": "", "type": "custom", "parameters_json": `[{"name":"api_key","value":"secret"}]`, "display_name": ""},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			names, blocks, err := processorsToState(tc.stateBlock, tc.processors)
			require.NoError(t, err)
			require.Equal(t, tc.expectNames, names)
			require.Equal(t, tc.expectBlocks, blocks)
		})
	}
}

func TestResourceConfigurationV2_inline(t *testing.T) {
	srv := fakebindplane.New(t)
	meta := newProviderMeta(t, srv, nil)
	r := resourceConfigurationV2()

	const (
		routerParams = `[{"name":"telemetry_types","value":["Logs"]},{"name":"routes","value":[{"id":"prod","condition":{"ottl":"attributes[\"env\"] == \"prod\""}}]}]`
		groupParams  = `[{"name":"telemetry_types","value":["Logs"]}]`
		batchParams  = `[{"name":"send_batch_size","value":200}]`
		customParams = `[{"name":"api_key","value":"secret"}]`
	)

	raw := map[string]any{
		"name":     "my-config",
		"platform": "linux",
		"rollout":  false,
		"source": []any{map[string]any{
			"name":      "my-source",
			"processor": []any{map[string]any{"type": "custom", "parameters_json": customParams}},
			"route": []any{
				map[string]any{"route_id": "source-logs", "telemetry_type": "logs", "components": []any{"connectors/router"}},
			},
		}},
		"connector": []any{map[string]any{
			"route_id":        "router",
			"type":            "routing",
			"parameters_json": routerParams,
			"route": []any{
				map[string]any{"route_id": "prod", "telemetry_type": "logs", "components": []any{"processors/pg-1"}},
			},
		}},
		"processor_group": []any{map[string]any{
			"route_id":        "pg-1",
			"parameters_json": groupParams,
			"processor": []any{
				map[string]any{"name": "my-filter"},
				map[string]any{"type": "batch", "display_name": "Batch", "parameters_json": batchParams},
			},
			"route": []any{
				map[string]any{"route_id": "pg-logs", "telemetry_type": "logs", "components": []any{"destinations/dest-1"}},
			},
		}},
		"destination": []any{map[string]any{
			"route_id":  "dest-1",
			"name":      "my-dest",
			"processor": []any{map[string]any{"name": "my-filter", "display_name": "Filter"}},
		}},
	}

	d := schema.TestResourceDataRaw(t, r.Schema, raw)
	require.NoError(t, r.Create(d, meta))
	requireGoldenJSON(t, "configuration_v2_inline_apply.json", lastApplyBody(t, srv))

	// Create reads the configuration back into state
	require.Equal(t, "", d.Get("connector.0.name"))
	require.Equal(t, "routing", d.Get("connector.0.type"))
	require.JSONEq(t, routerParams, d.Get("connector.0.parameters_json").(string))
	require.JSONEq(t, groupParams, d.Get("processor_group.0.parameters_json").(string))
	require.Empty(t, d.Get("processor_group.0.processors"))
	require.Equal(t, "my-filter", d.Get("processor_group.0.processor.0.name"))
	require.Equal(t, "batch", d.Get("processor_group.0.processor.1.type"))
	require.Equal(t, "Batch", d.Get("processor_group.0.processor.1.display_name"))
	require.JSONEq(t, batchParams, d.Get("processor_group.0.processor.1.parameters_json").(string))

	// Bindplane pins versions, assigns processor IDs, adds default
	// parameters, and masks sensitive parameters of source processors.
	srv.Seed(t, `{
		"apiVersion": "bindplane.observiq.com/v2",
		"kind": "Configuration",
		"metadata": {"id": "`+d.Id()+`", "name": "my-config", "version": 2, "labels": {"platform": "linux"}},
		"spec": {
			"contentType": "text/yaml",
			"sources": [{
				"id": "s-my-source",
				"name": "my-source:2",
				"processors": [{
					"id": "custom-1",
					"type": "custom:1",
					"parameters": [
						{"name": "api_key", "value": "(sensitive)", "sensitive": true},
						{"name": "enabled", "value": true}
					]
				}],
				"routes": {"logs": [{"id": "source-logs", "components": ["connectors/router"]}]}
			}],
			"connectors": [{
				"id": "router",
				"type": "routing:1",
				"parameters": [
					{"name": "telemetry_types", "value": ["Logs"]},
					{"name": "routes", "value": [{"id": "prod", "condition": {"ottl": "attributes[\"env\"] == \"prod\""}}]},
					{"name": "default_route", "value": ""}
				],
				"routes": {"logs": [{"id": "prod", "components": ["processors/pg-1"]}]}
			}],
			"processors": [{
				"id": "pg-1",
				"parameters": [{"name": "telemetry_types", "value": ["Logs"]}],
				"processors": [
					{"id": "p-my-filter", "name": "my-filter:4"},
					{
						"id": "batch-1",
						"displayName": "Batch",
						"type": "batch:3",
						"parameters": [
							{"name": "send_batch_size", "value": 200},
							{"name": "timeout", "value": "200ms"}
						]
					}
				],
				"routes": {"logs": [{"id": "pg-logs", "components": ["destinations/dest-1"]}]}
			}],
			"destinations": [{
				"id": "dest-1",
				"name": "my-dest:1",
				"processors": [{"id": "p-my-filter", "name": "my-filter:4", "displayName": "Filter"}]
			}],
			"selector": {"matchLabels": {"configuration": "my-config"}}
		}
	}`)
	require.NoError(t, r.Read(d, meta))

	require.Equal(t, "routing", d.Get("connector.0.type"))
	require.JSONEq(t, routerParams, d.Get("connector.0.parameters_json").(string))
	require.Equal(t, []any{
		map[string]any{"id": "p-my-filter", "name": "my-filter", "type": "", "parameters_json": "", "display_name": ""},
		map[string]any{"id": "batch-1", "name": "", "type": "batch", "parameters_json": batchParams, "display_name": "Batch"},
	}, d.Get("processor_group.0.processor"))
	require.Equal(t, []any{
		map[string]any{"id": "custom-1", "name": "", "type": "custom", "parameters_json": customParams, "display_name": ""},
	}, d.Get("source.0.processor"))
	require.Equal(t, []any{
		map[string]any{"id": "p-my-filter", "name": "my-filter", "type": "", "parameters_json": "", "display_name": "Filter"},
	}, d.Get("destination.0.processor"))

	// The configuration read from Bindplane matches the user's
	// configuration, so Terraform does not plan any changes.
	diff, err := r.SimpleDiff(context.Background(), d.State(), terraform.NewResourceConfigRaw(raw), meta)
	require.NoError(t, err)
	require.True(t, diff.Empty(), "unexpected diff: %v", diff)

	// Update sends the IDs Bindplane generated for inline processors
	require.NoError(t, r.Update(d, meta))
	var body struct {
		Resources []struct {
			Spec model.ConfigurationSpec `json:"spec"`
		} `json:"resources"`
	}
	require.NoError(t, json.Unmarshal([]byte(lastApplyBody(t, srv)), &body))
	require.Len(t, body.Resources, 1)
	spec := body.Resources[0].Spec
	require.Equal(t, "custom-1", spec.Sources[0].Processors[0].ID)
	require.Equal(t, "", spec.Processors[0].Processors[0].ID, "processors referenced by name are not sent with an ID")
	require.Equal(t, "batch-1", spec.Processors[0].Processors[1].ID)
}

func TestResourceConfigurationV2_inlineValidation(t *testing.T) {
	cases := []struct {
		name      string
		block     string
		value     map[string]any
		expectErr string
	}{
		{
			name:      "connector with name and type",
			block:     "connector",
			value:     map[string]any{"route_id": "router", "name": "my-router", "type": "routing"},
			expectErr: "connector 'router': name and type cannot both be set",
		},
		{
			name:      "connector without name or type",
			block:     "connector",
			value:     map[string]any{"route_id": "router"},
			expectErr: "connector 'router': one of name or type must be set",
		},
		{
			name:      "connector parameters without type",
			block:     "connector",
			value:     map[string]any{"route_id": "router", "name": "my-router", "parameters_json": `[]`},
			expectErr: "connector 'router': parameters_json requires type",
		},
		{
			name:      "processor group with invalid parameters",
			block:     "processor_group",
			value:     map[string]any{"route_id": "pg-1", "parameters_json": `{}`},
			expectErr: "processor group 'pg-1': failed to unmarshal parameters",
		},
		{
			name:  "processor group with processors and processor",
			block: "processor_group",
			value: map[string]any{
				"route_id":   "pg-1",
				"processors": []any{"my-filter"},
				"processor":  []any{map[string]any{"name": "my-batch"}},
			},
			expectErr: "processor group 'pg-1': processors and processor cannot both be set",
		},
		{
			name:  "source processor with name and type",
			block: "source",
			value: map[string]any{
				"name":      "my-source",
				"processor": []any{map[string]any{"name": "my-batch", "type": "batch"}},
			},
			expectErr: "source 'my-source': processor 0: name and type cannot both be set",
		},
		{
			name:  "destination processor without name or type",
			block: "destination",
			value: map[string]any{
				"route_id":  "dest-1",
				"name":      "my-dest",
				"processor": []any{map[string]any{"display_name": "Batch"}},
			},
			expectErr: "destination 'my-dest': processor 0: one of name or type must be set",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := fakebindplane.New(t)
			meta := newProviderMeta(t, srv, nil)
			r := resourceConfigurationV2()

			d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
				"name":     "my-config",
				"platform": "linux",
				"rollout":  false,
				tc.block:   []any{tc.value},
			})
			require.ErrorContains(t, r.Create(d, meta), tc.expectErr)
			require.Empty(t, srv.RequestsTo(http.MethodPost, "/v1/apply"))
		})
	}
}
