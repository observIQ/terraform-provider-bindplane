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
	"reflect"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/observiq/terraform-provider-bindplane/client"
)

// suppressDefaultParameterDiffs returns a DiffSuppressFunc for parameters_json
// that suppresses diffs which only remove parameters set to their resource
// type's default value.
//
// When a resource is applied, Bindplane saves a default value for every
// parameter that has one and was not set. Provider versions that saved every
// parameter returned by Bindplane to state left these defaults in state, and
// so do imports. Removing them from state does not change the resource,
// because Bindplane saves the same defaults again, so the diff is suppressed.
// Removing a parameter that is not at its default value is a real change, and
// is not suppressed.
//
// DiffSuppressFunc is not passed the provider meta, so meta is called to get
// the Bindplane client. If the client or the resource type cannot be
// retrieved, the diff is not suppressed.
func suppressDefaultParameterDiffs(rKind model.Kind, meta func() any) schema.SchemaDiffSuppressFunc {
	// Resource types are immutable for a given version, so parameter
	// definitions are cached by versioned type name.
	var mu sync.Mutex
	definitions := map[string][]model.ParameterDefinition{}

	return func(k, old, newVal string, d *schema.ResourceData) bool {
		if suppressEquivalentJSONDiffs(k, old, newVal, d) {
			return true
		}

		// A resource that does not exist yet has no state to compare against.
		if d.Id() == "" {
			return false
		}

		removed, ok := removedParameters(old, newVal)
		if !ok || len(removed) == 0 {
			return false
		}

		bindplane, ok := meta().(*client.BindPlane)
		if !ok || bindplane == nil {
			return false
		}

		// The resource's type includes the version of the resource type
		// it uses, such as "host:1". Defaults can differ between versions.
		g, err := bindplane.GenericResource(rKind, d.Get("name").(string))
		if err != nil || g == nil {
			return false
		}

		mu.Lock()
		defs, ok := definitions[g.Spec.Type]
		mu.Unlock()
		if !ok {
			defs, err = bindplane.ResourceTypeParameters(rKind.ResourceTypeKind(), g.Spec.Type)
			if err != nil {
				return false
			}
			mu.Lock()
			definitions[g.Spec.Type] = defs
			mu.Unlock()
		}

		return parametersAreDefaults(removed, defs)
	}
}

// removedParameters returns the parameters in old that are not in newVal,
// where old and newVal are serialized parameters_json values. ok is false if
// newVal adds or changes a parameter, or if either value cannot be parsed,
// because the diff is then more than a removal.
func removedParameters(old, newVal string) (removed []model.Parameter, ok bool) {
	oldParams, err := unmarshalParameters(old)
	if err != nil {
		return nil, false
	}

	newParams, err := unmarshalParameters(newVal)
	if err != nil {
		return nil, false
	}

	oldByName := make(map[string]model.Parameter, len(oldParams))
	for _, p := range oldParams {
		oldByName[p.Name] = p
	}

	newNames := make(map[string]bool, len(newParams))
	for _, p := range newParams {
		oldParam, exists := oldByName[p.Name]
		if !exists || !jsonEqual(oldParam, p) {
			return nil, false
		}
		newNames[p.Name] = true
	}

	for _, p := range oldParams {
		if !newNames[p.Name] {
			removed = append(removed, p)
		}
	}

	return removed, true
}

// parametersAreDefaults returns true if every parameter is set to the
// default value of its definition in defs. Parameters without a default are
// not saved by Bindplane when unset, so removing one is a real change.
func parametersAreDefaults(params []model.Parameter, defs []model.ParameterDefinition) bool {
	defaults := make(map[string]any, len(defs))
	for _, def := range defs {
		if def.Default != nil {
			defaults[def.Name] = def.Default
		}
	}

	for _, p := range params {
		def, ok := defaults[p.Name]
		if !ok || !jsonEqual(p.Value, def) {
			return false
		}
	}

	return true
}

func unmarshalParameters(s string) ([]model.Parameter, error) {
	params := []model.Parameter{}
	if s == "" {
		return params, nil
	}
	if err := json.Unmarshal([]byte(s), &params); err != nil {
		return nil, err
	}
	return params, nil
}

// jsonEqual compares a and b by their JSON representation, so that values
// decoded from different sources, such as state and the Bindplane API,
// compare equal when they serialize the same.
func jsonEqual(a, b any) bool {
	aBytes, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bBytes, err := json.Marshal(b)
	if err != nil {
		return false
	}

	var aData, bData any
	if err := json.Unmarshal(aBytes, &aData); err != nil {
		return false
	}
	if err := json.Unmarshal(bBytes, &bData); err != nil {
		return false
	}

	return reflect.DeepEqual(aData, bData)
}
