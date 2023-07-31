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
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/terraform-provider-bindplane/client"
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
	// re-create the resource or comfirm that it was deleted.
	if g == nil {
		d.SetId("")
		return nil
	}

	// Save values returned by bindplane to Terraform's state

	d.SetId(g.ID)

	// If the state ID is set but differs from the ID returned by,
	// bindplane, mark the resource to be re-created by unsetting
	// the ID. This will cause Terraform to attempt to create the resource
	// instead of updating it. The creation step will fail because
	// the resource already exists. This behavior is desirable, it will
	// prevent Terraform from modifying resources created by other means.
	if id := d.Id(); id != "" {
		if g.ID != d.Id() {
			d.SetId("")
			return nil
		}
	}

	if err := d.Set("name", g.Name); err != nil {
		return err
	}

	rType := strings.Split(g.Spec.Type, ":")[0]
	if err := d.Set("type", rType); err != nil {
		return err
	}

	// Prevent Terraform from attempting to update resources
	// with sensitive parameters. This is done by calculating
	// the parameters_applied option by writing (sensative)
	// to the state instead of the raw value. NOTE: parameters_json
	// will still contain the orional value provided by the user's
	// terraform configuration. Users should use an encrypted
	// backend if they wish to keep sensitive parameters out of plain
	// text.

	// Update param value's to "(sensitive)" if the parameter is sensitive
	paramsApplied := g.Spec.Parameters
	for i, p := range paramsApplied {
		if p.Sensitive {
			paramsApplied[i].Value = "(sensitive)"
		}
	}

	// Write paramsApplied to parameters_applied in the state
	paramStr, err := parameter.ParametersToString(paramsApplied)
	if err != nil {
		return err
	}
	return d.Set("parameters_applied", paramStr)
}

// genericResourceDelete can delete configurations, sources,
// destinations, and processors from the BindPlane API.
func genericResourceDelete(rKind model.Kind, d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)
	name := d.Get("name").(string)

	if err := bindplane.Delete(rKind, name); err != nil {
		return err
	}
	return resourceProcessorRead(d, meta)
}
