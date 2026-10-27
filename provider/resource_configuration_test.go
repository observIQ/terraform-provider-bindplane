// Copyright observIQ, Inc.
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadRolloutOptions(t *testing.T) {
	schemaMap := map[string]*schema.Schema{
		"rollout_options": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Required: true,
					},
					"parameters": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:     schema.TypeString,
									Required: true,
								},
								"value": {
									Type:     schema.TypeList,
									Required: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"labels": {
												Type:     schema.TypeMap,
												Required: true,
											},
											"name": {
												Type:     schema.TypeString,
												Required: true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, schemaMap, map[string]interface{}{
		"rollout_options": []interface{}{
			map[string]interface{}{
				"type": "progressive",
				"parameters": []interface{}{
					map[string]interface{}{
						"name": "stages",
						"value": []interface{}{
							map[string]interface{}{
								"labels": map[string]interface{}{
									"env": "stage",
								},
								"name": "stage",
							},
							map[string]interface{}{
								"labels": map[string]interface{}{
									"env": "production",
								},
								"name": "production",
							},
						},
					},
				},
			},
		},
	})

	expected := model.ResourceConfiguration{
		ParameterizedSpec: model.ParameterizedSpec{
			Type: "progressive",
			Parameters: []model.Parameter{
				{
					Name: "stages",
					Value: []interface{}{
						map[string]interface{}{
							"labels": map[string]interface{}{
								"env": "stage",
							},
							"name": "stage",
						},
						map[string]interface{}{
							"labels": map[string]interface{}{
								"env": "production",
							},
							"name": "production",
						},
					},
				},
			},
		},
	}

	resourceConfig, err := readRolloutOptions(resourceData)
	assert.NoError(t, err)
	assert.Equal(t, expected, resourceConfig)
}

func TestReadRolloutOptions_Empty(t *testing.T) {
	schemaMap := map[string]*schema.Schema{
		"rollout_options": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Required: true,
					},
					"parameters": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:     schema.TypeString,
									Required: true,
								},
								"value": {
									Type:     schema.TypeList,
									Required: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"labels": {
												Type:     schema.TypeMap,
												Required: true,
											},
											"name": {
												Type:     schema.TypeString,
												Required: true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, schemaMap, map[string]interface{}{})

	resourceConfig, err := readRolloutOptions(resourceData)
	assert.NoError(t, err)
	assert.Equal(t, resourceConfig, model.ResourceConfiguration{})
}

func TestResourceConfigurationV2RoleLabel(t *testing.T) {
	// The gateway role is set with the role=gateway label. Guard against
	// the role label being filtered the way platform is.
	f := newFakeBindplane(t)

	d := schema.TestResourceDataRaw(t, resourceConfigurationV2().Schema, map[string]any{
		"name":     "gw-config",
		"platform": "linux",
		"labels": map[string]any{
			"role": "gateway",
		},
	})
	require.NoError(t, resourceConfigurationV2Create(d, f.client()))
	require.Equal(t, "id-gw-config", d.Id())

	metadata := f.lastApplied()["metadata"].(map[string]any)
	require.Equal(t, map[string]any{
		"platform": "linux",
		"role":     "gateway",
	}, metadata["labels"])

	require.Equal(t, map[string]any{"role": "gateway"}, d.Get("labels"))
	require.Equal(t, "linux", d.Get("platform"))
}
