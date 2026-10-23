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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/observiq/terraform-provider-bindplane/internal/model"
	"github.com/observiq/terraform-provider-bindplane/internal/parameter"
)

// genericResourceRead can read source, destination, and processors
// from the BindPlane API and set them.
func genericResourceRead(rKind model.Kind, d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)
	resourceName := d.Get("name").(string)

	g, err := bindplane.GenericResource(rKind, resourceName)
	if err != nil {
		return err
	}

	// A nil return from GenericResource indicates that the resource
	// did not exist. Terraform read operations should always set the
	// ID to "" and return a nil error. This will allow Terraform to
	// re-create the resource or confirm that it was deleted.
	if g == nil {
		d.SetId("")
		return nil
	}

	// If the state ID is set but differs from the ID returned by,
	// bindplane, mark the resource to be re-created by unsetting
	// the ID. This will cause Terraform to attempt to create the resource
	// instead of updating it. The creation step will fail because
	// the resource already exists. This behavior is desirable, it will
	// prevent Terraform from modifying resources created by other means.
	if g.ID != d.Id() {
		d.SetId("")
		return nil
	}

	if err := d.Set("name", g.Name); err != nil {
		return err
	}

	rType := strings.Split(g.Spec.Type, ":")[0]
	if err := d.Set("type", rType); err != nil {
		return err
	}

	// Parameters defined by the user, previously saved to state
	stateParams := []model.Parameter{}
	if s := d.Get("parameters_json").(string); s != "" {
		if err := json.Unmarshal([]byte(s), &stateParams); err != nil {
			return fmt.Errorf("failed to unmarshal state paramters: %w", err)
		}
	}

	// Parameters returned by BindPlane API
	incomingParams := reconcileParameters(stateParams, g.Spec.Parameters)

	paramStr, err := parameter.ParametersToString(incomingParams)
	if err != nil {
		return err
	}

	return d.Set("parameters_json", paramStr)
}

// reconcileParameters returns the parameters that should be saved to state
// after reading a resource from the Bindplane API.
//
// Bindplane returns every parameter supported by the resource type, with
// defaults resolved. Only parameters that were previously saved to state,
// which mirror the user's configuration, are kept. This prevents parameters
// the user never configured from showing as a permanent diff. Parameters are
// returned in the same order as state so that the JSON comparison done by
// suppressEquivalentJSONDiffs is not affected by API ordering. Parameters in
// state that are missing from the API response are dropped, which surfaces
// them as drift.
func reconcileParameters(stateParams, incomingParams []model.Parameter) []model.Parameter {
	incomingByName := make(map[string]model.Parameter, len(incomingParams))
	for _, p := range incomingParams {
		incomingByName[p.Name] = p
	}

	params := make([]model.Parameter, 0, len(stateParams))
	for _, stateParam := range stateParams {
		incomingParam, ok := incomingByName[stateParam.Name]
		if !ok {
			continue
		}

		// Update sensitive parameters with the values from state
		// instead of saving "(sensitive value)" to state.
		if incomingParam.Sensitive {
			// Set the value to the value provided by the user in order to
			// prevent terraform from attempting to update the value.
			incomingParam.Value = stateParam.Value

			// Preserve the sensitive value to whatever the user configured, which
			// could be true, false, or nothing.
			incomingParam.Sensitive = stateParam.Sensitive
		}

		params = append(params, incomingParam)
	}

	return params
}

// importParameters returns the parameters that should be saved to state
// when importing a resource. All parameters are kept so that the imported
// state reflects the resource as it exists in Bindplane, except sensitive
// parameters, which Bindplane returns masked. The mask is not a usable value.
func importParameters(incomingParams []model.Parameter) []model.Parameter {
	params := make([]model.Parameter, 0, len(incomingParams))
	for _, p := range incomingParams {
		if p.Sensitive {
			continue
		}
		params = append(params, p)
	}
	return params
}

// genericResourceDelete can delete configurations, sources,
// destinations, and processors from the BindPlane API.
func genericResourceDelete(rKind model.Kind, d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)
	name := d.Get("name").(string)
	return bindplane.Delete(rKind, name)
}

// genericResourceImport imports a BindPlane resource by looking it up
// by its name.
func genericResourceImport(rKind model.Kind, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	bindplane := meta.(*client.BindPlane)

	// When importing, name is not set in the state so we need to grab
	// the ID instead, which is the same as "name".
	name := d.Id()

	g, err := bindplane.GenericResource(rKind, name)
	if err != nil {
		return nil, err
	}

	// bindplane.GenericResource will return a nil error if the resource
	// does not exist. It is up to the caller to check.
	if g == nil {
		return nil, fmt.Errorf("%s with name '%s' does not exist", rKind, name)
	}

	// Set the state ID to BindPlane's resource ID so that the next Read
	// does not clear the ID (genericResourceRead requires g.ID == d.Id()).
	d.SetId(g.ID)

	// Add the name to state, which will cause the import to succeed.
	if err := d.Set("name", g.Name); err != nil {
		return nil, fmt.Errorf("failed to set resource name in state for imported %s '%s': %v", rKind, name, err)
	}

	// Seed state with the resource's parameters. The following Read only keeps
	// parameters that exist in state, so without this an import would save no
	// parameters at all.
	paramStr, err := parameter.ParametersToString(importParameters(g.Spec.Parameters))
	if err != nil {
		return nil, err
	}
	if err := d.Set("parameters_json", paramStr); err != nil {
		return nil, fmt.Errorf("failed to set parameters in state for imported %s '%s': %v", rKind, name, err)
	}

	return []*schema.ResourceData{d}, nil
}
