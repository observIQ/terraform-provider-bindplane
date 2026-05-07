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
				Description: "The resource name for the fleet. This is used internally and cannot be changed after creation.",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A user-friendly name for the fleet that can be changed anytime.",
			},
			"agent_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The agent type (collector type) for agents in this fleet.",
			},
			"platform": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The platform for agents in this fleet.",
			},
			"configuration": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the configuration assigned to the fleet.",
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
	displayName := d.Get("display_name").(string)
	configuration := d.Get("configuration").(string)
	agentType := d.Get("agent_type").(string)
	platform := d.Get("platform").(string)

	if name == "" {
		return fmt.Errorf("a fleet name is required in order to create")
	}

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

		// Use name as ID (not a random UUID)
		d.SetId(name)
	}

	// Build labels with auto-generated agent-type and platform
	labelsMapStr := make(map[string]string)
	if agentType != "" {
		labelsMapStr["agent-type"] = agentType
	}
	if platform != "" {
		labelsMapStr["platform"] = platform
	}
	labels := model.LabelsFromValidatedMap(labelsMapStr)

	// Auto-generate selector matching the fleet name
	selector := model.AgentSelector{
		MatchLabels: map[string]string{
			"fleet": name,
		},
	}

	// Build FleetSpec from schema
	fleetSpec := model.FleetSpec{
		Configuration: configuration,
		Selector:      selector,
	}

	// Convert FleetSpec to map[string]any
	var specMap map[string]any
	if err := mapstructure.Decode(fleetSpec, &specMap); err != nil {
		return fmt.Errorf("failed to encode fleet spec: %w", err)
	}

	// Create AnyResource
	anyResource := &model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       model.KindFleet,
			Metadata: model.Metadata{
				ID:          name,
				Name:        name,
				DisplayName: displayName,
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

	// Set display name if available
	if fleet.Metadata.DisplayName != "" {
		if err := d.Set("display_name", fleet.Metadata.DisplayName); err != nil {
			return err
		}
	}

	// Set agent-type and platform from labels
	labelsMap := fleet.GetLabels()
	if agentType, ok := labelsMap.Set["agent-type"]; ok {
		if err := d.Set("agent_type", agentType); err != nil {
			return err
		}
	}
	if platform, ok := labelsMap.Set["platform"]; ok {
		if err := d.Set("platform", platform); err != nil {
			return err
		}
	}

	// Set configuration
	return d.Set("configuration", fleet.FleetSpec.Configuration)
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
