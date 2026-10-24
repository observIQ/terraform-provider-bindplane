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

package client

import (
	"github.com/observiq/bindplane-op-enterprise/model"
)

// This file contains the request and response bodies of the Bindplane
// REST API endpoints used by the client. They contain only the fields
// used by the provider.

// applyRequest is the body of POST /v1/apply
type applyRequest struct {
	Resources []*model.AnyResource `json:"resources"`
}

// applyResponse is the response to POST /v1/apply
type applyResponse struct {
	Updates []*model.AnyResourceStatus `json:"updates"`
}

// startRolloutRequest is the body of POST /v1/rollouts/{name}/start
type startRolloutRequest struct {
	// Options is always null, so Bindplane uses the rollout
	// options of the configuration
	Options any `json:"options"`
}

// configurationResponse is the response to GET /v1/configurations/{name}
type configurationResponse struct {
	Configuration *model.Configuration `json:"configuration"`
}

// sourceResponse is the response to GET /v1/sources/{name}
type sourceResponse struct {
	Source *model.Source `json:"source"`
}

// destinationResponse is the response to GET /v1/destinations/{name}
type destinationResponse struct {
	Destination *model.Destination `json:"destination"`
}

// processorResponse is the response to GET /v1/processors/{name}
type processorResponse struct {
	Processor *model.Processor `json:"processor"`
}

// extensionResponse is the response to GET /v1/extensions/{name}
type extensionResponse struct {
	Extension *model.Extension `json:"extension"`
}

// connectorResponse is the response to GET /v1/resources/Connector/{name}
type connectorResponse struct {
	Resource *model.Connector `json:"resource"`
}

// fleetResponse is the response to GET /v1/resources/Fleet/{name}
type fleetResponse struct {
	Resource *model.Fleet `json:"resource"`
}

// gatewayResponse is the response to GET /v1/resources/Gateway/{name}
type gatewayResponse struct {
	Resource *model.Gateway `json:"resource"`
}

// errorResponse is the body of an unsuccessful response
type errorResponse struct {
	Errors []string `json:"errors"`
}
