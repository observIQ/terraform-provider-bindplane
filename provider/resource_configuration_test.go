package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/stretchr/testify/assert"
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
