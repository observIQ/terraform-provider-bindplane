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

// Package resource provides functions for defining bindplane
// generic resources.
package resource

import "github.com/observiq/bindplane-op/model"

// AnyResourceFromConfiguration takes a BindPlane configuration and returns a BindPlane AnyResource
func AnyResourceFromConfiguration(c *model.Configuration) model.AnyResource {
	a := model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: c.APIVersion,
			Kind:       c.Kind,
			Metadata:   c.Metadata,
		},
		Spec: map[string]any{
			"contentType": c.Spec.ContentType,
			"selector":    c.Spec.Selector,
		},
	}

	if c.Spec.Raw != "" {
		a.Spec["raw"] = c.Spec.Raw
	} else {
		a.Spec["sources"] = c.Spec.Sources
		a.Spec["destinations"] = c.Spec.Destinations
	}

	return a
}
