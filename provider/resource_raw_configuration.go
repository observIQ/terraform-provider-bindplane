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
	"errors"
	"fmt"
	"time"

	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/observiq/terraform-provider-bindplane/internal/configuration"
	"github.com/observiq/terraform-provider-bindplane/internal/maputil"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRawConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceRawConfigurationCreate,
		Update: resourceRawConfigurationCreate, // Run create as update
		Read:   resourceRawConfigurationRead,
		Delete: genericConfigurationDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the configuration.",
			},
			"platform": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(val any, _ string) (warns []string, errs []error) {
					platform := val.(string)
					if !isValidPlatform(platform) {
						errs = append(errs, fmt.Errorf("%s is not a valid platform", platform))
					}
					return nil, errs
				},
				Description: "The platform the configuration is for.",
			},
			"labels": {
				Type:     schema.TypeMap,
				Required: true,
				ForceNew: false,
				ValidateFunc: func(val any, _ string) (warns []string, errs []error) {
					labels := val.(map[string]any)
					_, ok := labels["platform"]
					if ok {
						errs = append(errs, errors.New("label 'platform' will be overwritten by the configured platform"))
					}
					return
				},
				Description: "Key value pairs which will be added to the configuration as labels.",
			},
			"match_labels": {
				Type:        schema.TypeMap,
				Computed:    true,
				ForceNew:    false,
				Description: "Labels that BindPlane OP uses to determine which agents the configuration should apply to. This value is computed by Terraform and is not user configurable.",
			},
			"raw_configuration": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "The OpenTelemetry collector configuration to use for the raw configuration.",
			},
			"rollout": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "Whether or not to trigger a rollout automatically when a configuration is updated. When set to true, BindPlane OP will automatically roll out the configuration change to managed agents.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(maxTimeout),
			Read:   schema.DefaultTimeout(maxTimeout),
			Delete: schema.DefaultTimeout(maxTimeout),
		},
	}
}

func resourceRawConfigurationCreate(d *schema.ResourceData, meta any) error {
	name := d.Get("name").(string)
	rollout := d.Get("rollout").(bool)

	labels, err := maputil.StringMapFromTFMap(d.Get("labels").(map[string]any))
	if err != nil {
		return err
	}
	labels["platform"] = d.Get("platform").(string)

	// Match labels should always be configuration=<name>
	matchLabels := map[string]string{
		"configuration": name,
	}

	opts := []configuration.Option{
		configuration.WithName(name),
		configuration.WithLabels(labels),
		configuration.WithMatchLabels(matchLabels),
		configuration.WithRawOTELConfig(d.Get("raw_configuration").(string)),
	}

	config, err := configuration.NewV1(opts...)
	if err != nil {
		return fmt.Errorf("failed to create new raw configuration: %v", err)
	}

	resource := resource.AnyResourceFromRawConfigurationV1(config)
	bindplane := meta.(*client.BindPlane)
	ctx := context.TODO()
	timeout := d.Timeout(schema.TimeoutCreate) - time.Minute
	if err := bindplane.ApplyWithRetry(ctx, timeout, &resource, rollout); err != nil {
		return err
	}

	return resourceConfigurationRead(d, meta)
}

func resourceRawConfigurationRead(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	config, err := bindplane.Configuration(d.Get("name").(string))
	if err != nil {
		return err
	}

	if config == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("name", config.Name()); err != nil {
		return err
	}

	labels := config.Metadata.Labels.AsMap()
	platform, ok := labels["platform"]
	if ok {
		if err := d.Set("platform", platform); err != nil {
			return err
		}
		// Remove the platform label from the labels map
		// because Terraform's state does not expect it.
		delete(labels, "platform")
	}

	// Save the labels map to state, which has the 'platform' label removed.
	if err := d.Set("labels", labels); err != nil {
		return err
	}

	if err := d.Set("raw_configuration", config.Spec.Raw); err != nil {
		return err
	}

	matchLabels := make(map[string]string)
	for k, v := range config.Spec.Selector.MatchLabels {
		matchLabels[k] = v
	}
	if err := d.Set("match_labels", matchLabels); err != nil {
		return err
	}

	d.SetId(config.ID())
	return nil
}
