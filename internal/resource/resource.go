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

// AnyResourceFromConfiguration takes an BindPlane configuration and returns an BindPlane AnyResource
func AnyResourceFromConfiguration(c *model.Configuration) model.AnyResource {
	return model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: c.APIVersion,
			Kind:       c.Kind,
			Metadata:   c.Metadata,
		},
		Spec: map[string]any{
			"contentType": c.Spec.ContentType,
			"raw":         c.Spec.Raw,
			"selector":    c.Spec.Selector,
		},
	}
}

// AnyResourceFromDestination takes an BindPlane Destination and returns an BindPlane AnyResource
func AnyResourceFromDestination(d *model.Destination) model.AnyResource {
	r := model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: d.APIVersion,
			Kind:       d.Kind,
			Metadata:   d.Metadata,
		},
		Spec: map[string]any{
			"type":       "googlecloud",
			"parameters": []map[string]any{},
		},
	}

	params := []map[string]any{}
	for _, p := range d.Spec.Parameters {
		param := map[string]any{}
		param[p.Name] = p.Value
		params = append(params, param)
	}

	r.Spec["parameters"] = params

	return r
}
