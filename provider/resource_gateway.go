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
	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/observiq/terraform-provider-bindplane/internal/model"
)

func resourceGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceGatewayCreate,
		Update: resourceGatewayCreate,
		Read:   resourceGatewayRead,
		Delete: resourceGatewayDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGatewayImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The resource name for the gateway. Collectors join the gateway with the label 'gateway=<name>'. Cannot be changed after creation.",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A user-friendly name for the gateway that can be changed anytime.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the gateway.",
			},
			"configuration": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateGatewayConfiguration,
				Description:  "Name of the configuration assigned to the gateway's collectors, without a version.",
			},
			"load_balancer_hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Hostname of the load balancer in front of the gateway's collectors.",
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

// validateGatewayConfiguration rejects versioned configuration references
// such as "my-config:2". Bindplane rejects them on apply because the live
// version is controlled by rollouts.
func validateGatewayConfiguration(val any, key string) (warns []string, errs []error) {
	configuration := val.(string)
	if model.TrimVersion(configuration) != configuration {
		errs = append(errs, fmt.Errorf("%s must be a configuration name without a version, got '%s'", key, configuration))
	}
	return warns, errs
}

func resourceGatewayCreate(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	name := d.Get("name").(string)
	configuration := d.Get("configuration").(string)

	// Bindplane does not validate the gateway's configuration reference,
	// so check that it exists to catch typos.
	if configuration != "" {
		cfg, err := bindplane.Configuration(configuration)
		if err != nil {
			return err
		}
		if cfg == nil {
			return fmt.Errorf("configuration '%s' does not exist", configuration)
		}
	}

	if d.Id() == "" {
		g, err := bindplane.Gateway(name)
		if err != nil {
			return err
		}
		if g != nil {
			return fmt.Errorf("gateway with name '%s' already exists", name)
		}

		// Use name as ID (not a random UUID)
		d.SetId(name)
	}

	// Only manual mode fields are sent. Never send spec.workload, which
	// makes the gateway operator mode, or status, which holds the
	// gateway's setup checklist in the UI.
	anyResource := &model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       model.KindGateway,
			Metadata: model.Metadata{
				ID:          d.Id(),
				Name:        name,
				DisplayName: d.Get("display_name").(string),
				Description: d.Get("description").(string),
			},
		},
		Spec: map[string]any{
			"configuration":        configuration,
			"loadBalancerHostname": d.Get("load_balancer_hostname").(string),
		},
	}

	ctx := context.Background()
	timeout := d.Timeout(schema.TimeoutCreate) - time.Minute
	if err := bindplane.ApplyWithRetry(ctx, timeout, anyResource, false); err != nil {
		return client.GatewayUnsupportedError(err)
	}

	return resourceGatewayRead(d, meta)
}

func resourceGatewayRead(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)
	name := d.Get("name").(string)

	gateway, err := bindplane.Gateway(name)
	if err != nil {
		return err
	}

	// If resource not found, clear the ID to allow recreation
	if gateway == nil {
		d.SetId("")
		return nil
	}

	// If the ID differs, the gateway was replaced outside of Terraform.
	if gateway.ID() != d.Id() {
		d.SetId("")
		return nil
	}

	if err := checkGatewayManualMode(gateway); err != nil {
		return err
	}

	if err := d.Set("name", gateway.Name()); err != nil {
		return err
	}

	// Always set every field, including empty values, so changes made
	// in Bindplane show up as drift.
	if err := d.Set("display_name", gateway.Metadata.DisplayName); err != nil {
		return err
	}
	if err := d.Set("description", gateway.Metadata.Description); err != nil {
		return err
	}
	if err := d.Set("configuration", gateway.GatewaySpec.Configuration); err != nil {
		return err
	}
	return d.Set("load_balancer_hostname", gateway.GatewaySpec.LoadBalancerHostname)
}

func resourceGatewayDelete(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)
	name := d.Get("name").(string)

	if err := bindplane.DeleteGateway(name); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func resourceGatewayImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	bindplane := meta.(*client.BindPlane)

	name := d.Id()

	gateway, err := bindplane.Gateway(name)
	if err != nil {
		return nil, err
	}

	if gateway == nil {
		return nil, fmt.Errorf("gateway with name '%s' does not exist", name)
	}

	if err := checkGatewayManualMode(gateway); err != nil {
		return nil, err
	}

	d.SetId(gateway.ID())

	if err := d.Set("name", gateway.Name()); err != nil {
		return nil, fmt.Errorf("failed to set gateway name: %w", err)
	}

	return []*schema.ResourceData{d}, nil
}

// checkGatewayManualMode returns an error for operator mode gateways,
// which are managed by the Bindplane Kubernetes operator.
func checkGatewayManualMode(gateway *model.Gateway) error {
	if gateway.Mode() != model.GatewayModeManual {
		return fmt.Errorf("gateway '%s' is managed by the Bindplane Kubernetes operator and is not supported by this resource", gateway.Name())
	}
	return nil
}
