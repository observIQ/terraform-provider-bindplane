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

import (
	"fmt"

	"github.com/observiq/bindplane-op/model"
)

// AnyResourceV1 takes a BindPlane resource name, kind, type, parameters
// and returns a bindplane.observiq.com/v1.AnyResource. Supported resources
// are Sources, Destinations, and Processors. For configurations, use
// AnyResourceFromConfigurationV1.
func AnyResourceV1(rName, rType string, rKind model.Kind, rParameters []model.Parameter) (model.AnyResource, error) {
	switch rKind {
	case model.KindSource, model.KindDestination, model.KindProcessor:
		return model.AnyResource{
			ResourceMeta: model.ResourceMeta{
				APIVersion: "bindplane.observiq.com/v1",
				Kind:       rKind,
				Metadata: model.Metadata{
					Name: rName,
				},
			},
			Spec: map[string]any{
				"type":       rType,
				"parameters": rParameters,
			},
		}, nil
	default:
		return model.AnyResource{}, fmt.Errorf("unknown bindplane resource kind: %s", rKind)
	}
}

// AnyResourceFromConfiguration takes a BindPlane configuration and returns a
// bindplane.observiq.com/v1.AnyResource
func AnyResourceFromConfigurationV1(c *model.Configuration) model.AnyResource {
	a := anyResourceFromConfiguration(c)

	if len(c.Spec.Sources) > 0 {
		a.Spec["sources"] = c.Spec.Sources
	}

	if len(c.Spec.Destinations) > 0 {
		a.Spec["destinations"] = c.Spec.Destinations
	}

	return a
}

// AnyResourceFromRawConfigurationV1 takes a BindPlane raw configuration and returns a BindPlane AnyResource
func AnyResourceFromRawConfigurationV1(c *model.Configuration) model.AnyResource {
	a := anyResourceFromConfiguration(c)
	a.Spec["raw"] = c.Spec.Raw
	return a
}

func anyResourceFromConfiguration(c *model.Configuration) model.AnyResource {
	return model.AnyResource{
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
}
