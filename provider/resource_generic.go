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
	"github.com/observiq/terraform-provider-bindplane/internal/client"
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

	if err := d.Set("name", g.Name); err != nil {
		return err
	}

	if err := d.Set("version", g.Version); err != nil {
		return err
	}

	rType := strings.Split(g.Spec.Type, ":")[0]
	if err := d.Set("type", rType); err != nil {
		return err
	}

	paramStr, err := parameter.ParametersToString(g.Spec.Parameters)
	if err != nil {
		return err
	}
	return d.Set("parameters_json", paramStr)
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
