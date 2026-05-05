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
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/mapstructure"
	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/observiq/terraform-provider-bindplane/internal/component"
)

func resourceFleet() *schema.Resource {
	return &schema.Resource{
		Create: resourceFleetCreate,
		Update: resourceFleetCreate,
		Read:   resourceFleetRead,
		Delete: resourceFleetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFleetImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the fleet.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the fleet.",
			},
			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Labels of the fleet.",
			},
			"configuration": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the configuration assigned to the fleet.",
			},
			"selector": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Agent selector for matching agents to this fleet.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"match_labels": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Labels for matching agents.",
						},
					},
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(maxTimeout),
			Read:   schema.DefaultTimeout(maxTimeout),
			Update: schema.DefaultTimeout(maxTimeout),
			Delete: schema.DefaultTimeout(maxTimeout),
		},
	}
}

func resourceFleetCreate(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	configuration := d.Get("configuration").(string)

	// Validate that configuration exists if provided
	if configuration != "" {
		cfg, err := bindplane.Configuration(configuration)
		if err != nil {
			return err
		}
		if cfg == nil {
			return fmt.Errorf("configuration '%s' does not exist", configuration)
		}
	}

	// Check if fleet already exists
	if d.Id() == "" {
		f, err := bindplane.Fleet(name)
		if err != nil {
			return err
		}
		if f != nil {
			return fmt.Errorf("fleet with name '%s' already exists", name)
		}

		// Generate a new ID
		d.SetId(component.NewResourceID())
	}

	// Build FleetSpec from schema
	fleetSpec := model.FleetSpec{
		Configuration: configuration,
		Selector:      buildAgentSelector(d),
	}

	// Convert FleetSpec to map[string]any
	var specMap map[string]any
	if err := mapstructure.Decode(fleetSpec, &specMap); err != nil {
		return fmt.Errorf("failed to encode fleet spec: %w", err)
	}

	// Build labels
	var labels model.Labels
	if labelsData, ok := d.GetOk("labels"); ok {
		labelsMap := labelsData.(map[string]interface{})
		labelsMapStr := make(map[string]string)
		for k, v := range labelsMap {
			labelsMapStr[k] = v.(string)
		}
		labels = model.LabelsFromValidatedMap(labelsMapStr)
	}

	// Create AnyResource
	anyResource := &model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       model.KindFleet,
			Metadata: model.Metadata{
				ID:          d.Id(),
				Name:        name,
				Description: description,
				Labels:      labels,
			},
		},
		Spec: specMap,
	}

	ctx := context.Background()
	timeout := d.Timeout(schema.TimeoutCreate) - time.Minute
	if err := bindplane.ApplyWithRetry(ctx, timeout, anyResource, false); err != nil {
		return err
	}

	return resourceFleetRead(d, meta)
}

func resourceFleetRead(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)
	name := d.Get("name").(string)

	fleet, err := bindplane.Fleet(name)
	if err != nil {
		return err
	}

	// If resource not found, clear the ID to allow recreation
	if fleet == nil {
		d.SetId("")
		return nil
	}

	// Validate ID
	if fleet.ID() != d.Id() {
		d.SetId("")
		return nil
	}

	// Set basic fields
	if err := d.Set("name", fleet.Name()); err != nil {
		return err
	}

	if err := d.Set("description", fleet.Description()); err != nil {
		return err
	}

	// Set labels
	labelsMap := fleet.GetLabels()
	labelsSetMap := make(map[string]string)
	for k, v := range labelsMap.Set {
		labelsSetMap[k] = v
	}
	if err := d.Set("labels", labelsSetMap); err != nil {
		return err
	}

	// Set configuration
	if err := d.Set("configuration", fleet.FleetSpec.Configuration); err != nil {
		return err
	}

	// Set selector
	selector := buildSelectorSchema(fleet.FleetSpec.Selector)
	return d.Set("selector", selector)
}

func resourceFleetDelete(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)
	name := d.Get("name").(string)

	if err := bindplane.DeleteFleet(name); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// buildAgentSelector creates an AgentSelector from terraform schema data
func buildAgentSelector(d *schema.ResourceData) model.AgentSelector {
	selector := model.AgentSelector{}

	if selectorData, ok := d.GetOk("selector"); ok {
		selectorList := selectorData.([]interface{})
		if len(selectorList) > 0 && selectorList[0] != nil {
			selectorMap := selectorList[0].(map[string]interface{})

			if matchLabels, ok := selectorMap["match_labels"]; ok {
				matchLabelsMap := matchLabels.(map[string]interface{})
				selector.MatchLabels = make(map[string]string)
				for k, v := range matchLabelsMap {
					selector.MatchLabels[k] = v.(string)
				}
			}
		}
	}

	return selector
}

// buildSelectorSchema converts a model.AgentSelector to terraform schema
func buildSelectorSchema(selector model.AgentSelector) []interface{} {
	if len(selector.MatchLabels) == 0 {
		return nil
	}

	selectorMap := map[string]interface{}{
		"match_labels": selector.MatchLabels,
	}

	return []interface{}{selectorMap}
}

func resourceFleetImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	bindplane := meta.(*client.BindPlane)

	name := d.Id()

	fleet, err := bindplane.Fleet(name)
	if err != nil {
		return nil, err
	}

	if fleet == nil {
		return nil, fmt.Errorf("fleet with name '%s' does not exist", name)
	}

	d.SetId(fleet.ID())

	if err := d.Set("name", fleet.Name()); err != nil {
		return nil, fmt.Errorf("failed to set fleet name: %w", err)
	}

	return []*schema.ResourceData{d}, nil
}
