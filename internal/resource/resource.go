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

// Package resource provides functions for defining Bindplane
// generic resources.
package resource

import (
	"fmt"

	"github.com/observiq/bindplane-op-enterprise/model"
)

// AnyResourceV1 requires a BindPlane resource name, kind, type, and parameters. Optionally
// takes a display name and description. Returns a bindplane.observiq.com/v1.AnyResource.
// Supported resources are Sources, Destinations, and Processors. For configurations, use
// AnyResourceFromConfigurationV1.
//
// rParameters and rProcessors can be nil.
func AnyResourceV1(id, rName, rType string, rKind model.Kind, rParameters []model.Parameter, rProcessors []model.ResourceConfiguration, displayName, description string) (model.AnyResource, error) {
	procs := []map[string]string{}
	for _, p := range rProcessors {
		proc := map[string]string{}

		if p.Name != "" {
			proc["name"] = p.Name
		}

		procs = append(procs, proc)
	}

	switch rKind {
	case model.KindSource, model.KindDestination, model.KindProcessor, model.KindExtension, model.KindConnector:
		r := model.AnyResource{
			ResourceMeta: model.ResourceMeta{
				APIVersion: "bindplane.observiq.com/v1",
				Kind:       rKind,
				Metadata: model.Metadata{
					ID:   id,
					Name: rName,
				},
			},
			Spec: map[string]any{
				"type":       rType,
				"parameters": rParameters,
				"processors": procs,
			},
		}

		if displayName != "" {
			r.Metadata.DisplayName = displayName
		}

		if description != "" {
			r.Metadata.Description = description
		}

		return r, nil
	default:
		return model.AnyResource{}, fmt.Errorf("unknown bindplane resource kind: %s", rKind)
	}
}

// AnyResourceFromConfigurationV1 takes a Bindplane configuration and returns a
// bindplane.observiq.com/v1.AnyResource
func AnyResourceFromConfigurationV1(c *model.Configuration) model.AnyResource {
	a := anyResourceFromConfiguration(c)

	if len(c.Spec.Sources) > 0 {
		a.Spec["sources"] = c.Spec.Sources
	}

	if len(c.Spec.Destinations) > 0 {
		a.Spec["destinations"] = c.Spec.Destinations
	}

	if len(c.Spec.Extensions) > 0 {
		a.Spec["extensions"] = c.Spec.Extensions
	}

	if c.Spec.Rollout.Type != "" {
		a.Spec["rollout"] = c.Spec.Rollout
	}

	if len(c.Spec.MeasurementInterval) > 0 {
		a.Spec["measurementInterval"] = c.Spec.MeasurementInterval
	}

	if len(c.Spec.Processors) > 0 {
		a.Spec["processors"] = c.Spec.Processors
	}

	if len(c.Spec.Connectors) > 0 {
		a.Spec["connectors"] = c.Spec.Connectors
	}

	if len(c.Spec.Parameters) > 0 {
		a.Spec["parameters"] = c.Spec.Parameters
	}

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
