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
	"fmt"
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

	var id string
	var name string
	var version model.Version
	var spec model.ParameterizedSpec

	switch rKind {
	case model.KindDestination:
		r, err := bindplane.Destination(resourceName)
		if err != nil {
			return err
		}

		// When a resource does not exist, always set the id to "" and return a nil error.
		if r == nil {
			d.SetId("")
			return nil
		}

		id = r.ID()
		name = r.Name()
		version = r.Version()
		spec = r.Spec
	case model.KindSource:
		r, err := bindplane.Source(resourceName)
		if err != nil {
			return err
		}

		// When a resource does not exist, always set the id to "" and return a nil error.
		if r == nil {
			d.SetId("")
			return nil
		}

		id = r.ID()
		name = r.Name()
		version = r.Version()
		spec = r.Spec
	case model.KindProcessor:
		r, err := bindplane.Processor(resourceName)
		if err != nil {
			return err
		}

		// When a resource does not exist, always set the id to "" and return a nil error.
		if r == nil {
			d.SetId("")
			return nil
		}

		id = r.ID()
		name = r.Name()
		version = r.Version()
		spec = r.Spec

	default:
		return fmt.Errorf("genericResourceRead does not support bindplane kind '%s'", rKind)
	}

	// Save values returned by bindplane to Terraform's state

	d.SetId(id)

	if err := d.Set("name", name); err != nil {
		return err
	}

	if err := d.Set("version", version); err != nil {
		return err
	}

	rType := strings.Split(spec.Type, ":")[0]
	if err := d.Set("type", rType); err != nil {
		return err
	}

	paramStr, err := parameter.ParametersToString(spec.Parameters)
	if err != nil {
		return err
	}
	if err := d.Set("parameters_json", paramStr); err != nil {
		return err
	}

	return nil
}

// genericResourceDelete can delete configurations, sources,
// destinations, and processors from the BindPlane API.
func genericResourceDelete(rKind model.Kind, d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)
	name := d.Get("name").(string)

	var err error
	switch rKind {
	case model.KindConfiguration:
		err = bindplane.DeleteConfiguration(name)
	case model.KindDestination:
		err = bindplane.DeleteDestination(name)
	case model.KindSource:
		err = bindplane.DeleteSource(name)
	case model.KindProcessor:
		err = bindplane.DeleteProcessor(name)
	default:
		return fmt.Errorf("genericResourceDelete does not support bindplane kind '%s'", rKind)
	}

	if err != nil {
		return err
	}
	return resourceProcessorRead(d, meta)
}
