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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/observiq/terraform-provider-bindplane/internal/model"
	"github.com/stretchr/testify/require"
)

// fakeBindplane is an in-memory Bindplane API serving the endpoints used by
// the gateway and configuration resources: POST /v1/apply,
// GET and DELETE /v1/resources/:kind/:name and GET /v1/configurations/:name.
type fakeBindplane struct {
	t      *testing.T
	server *httptest.Server

	mu sync.Mutex
	// resources holds stored resources, keyed by kind and name.
	resources map[model.Kind]map[string]*model.AnyResource
	// applied holds the raw JSON body of each resource sent to /v1/apply.
	applied []map[string]any
	// failures maps a request path to a response returned instead of
	// handling the request.
	failures map[string]fakeFailure
	// applyStatus, when set, is returned as the status of every applied
	// resource instead of "created".
	applyStatus model.UpdateStatus
}

// fakeFailure is an error response returned for a request path.
type fakeFailure struct {
	status int
	body   string
}

// fakeApplyPayload is the body of POST /v1/apply, decoded as raw JSON so
// tests can assert on exactly which fields were sent.
type fakeApplyPayload struct {
	Resources []json.RawMessage `json:"resources"`
}

// fakeApplyResponse is the body of the response to POST /v1/apply.
type fakeApplyResponse struct {
	Updates []*model.AnyResourceStatus `json:"updates"`
}

// fakeResourceResponse is the body of GET /v1/resources/:kind/:name.
type fakeResourceResponse struct {
	Resource *model.AnyResource `json:"resource"`
}

// fakeConfigurationResponse is the body of GET /v1/configurations/:name.
type fakeConfigurationResponse struct {
	Configuration *model.AnyResource `json:"configuration"`
}

func newFakeBindplane(t *testing.T) *fakeBindplane {
	f := &fakeBindplane{
		t:         t,
		resources: map[model.Kind]map[string]*model.AnyResource{},
		failures:  map[string]fakeFailure{},
	}
	f.server = httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(f.server.Close)
	return f
}

// client returns a provider client configured for the fake server.
func (f *fakeBindplane) client() *client.BindPlane {
	d := schema.TestResourceDataRaw(f.t, Configure().Schema, map[string]any{
		"remote_url": f.server.URL,
		"username":   "u",
		"password":   "p",
	})
	meta, diags := providerConfigure(d, nil)
	require.False(f.t, diags.HasError())
	return meta.(*client.BindPlane)
}

// put stores a resource as though it had been created in Bindplane.
func (f *fakeBindplane) put(r *model.AnyResource) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.resources[r.Kind] == nil {
		f.resources[r.Kind] = map[string]*model.AnyResource{}
	}
	f.resources[r.Kind][r.Metadata.Name] = r
}

// get returns a stored resource, or nil.
func (f *fakeBindplane) get(kind model.Kind, name string) *model.AnyResource {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.resources[kind][name]
}

// fail makes requests to path return status.
func (f *fakeBindplane) fail(path string, status int) {
	f.failWithBody(path, status, "")
}

// failWithBody makes requests to path return status and a JSON body.
func (f *fakeBindplane) failWithBody(path string, status int, body string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failures[path] = fakeFailure{status: status, body: body}
}

// lastApplied returns the raw JSON of the last resource sent to /v1/apply.
func (f *fakeBindplane) lastApplied() map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	require.NotEmpty(f.t, f.applied)
	return f.applied[len(f.applied)-1]
}

func (f *fakeBindplane) handle(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	failure, failed := f.failures[r.URL.Path]
	f.mu.Unlock()
	if failed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(failure.status)
		_, _ = w.Write([]byte(failure.body))
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/"), "/")
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/v1/apply":
		f.handleApply(w, r)
	case len(parts) == 3 && parts[0] == "resources":
		f.handleResource(w, r, model.Kind(parts[1]), parts[2])
	case r.Method == http.MethodGet && len(parts) == 2 && parts[0] == "configurations":
		resource := f.get(model.KindConfiguration, parts[1])
		if resource == nil {
			http.NotFound(w, r)
			return
		}
		f.writeJSON(w, fakeConfigurationResponse{Configuration: resource})
	default:
		http.NotFound(w, r)
	}
}

func (f *fakeBindplane) handleApply(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	require.NoError(f.t, err)

	var payload fakeApplyPayload
	require.NoError(f.t, json.Unmarshal(body, &payload))

	updates := []*model.AnyResourceStatus{}
	for _, raw := range payload.Resources {
		var fields map[string]any
		require.NoError(f.t, json.Unmarshal(raw, &fields))

		resource := &model.AnyResource{}
		require.NoError(f.t, json.Unmarshal(raw, resource))
		// Bindplane assigns an ID when none is sent
		if resource.Metadata.ID == "" {
			resource.Metadata.ID = "id-" + resource.Metadata.Name
		}

		f.mu.Lock()
		f.applied = append(f.applied, fields)
		applyStatus := f.applyStatus
		f.mu.Unlock()

		if applyStatus == "" {
			applyStatus = model.StatusCreated
			f.put(resource)
		}
		updates = append(updates, &model.AnyResourceStatus{
			Resource: *resource,
			Status:   applyStatus,
			Reason:   "fake reason",
		})
	}

	f.writeJSON(w, fakeApplyResponse{Updates: updates})
}

func (f *fakeBindplane) handleResource(w http.ResponseWriter, r *http.Request, kind model.Kind, name string) {
	resource := f.get(kind, name)
	if resource == nil {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		f.writeJSON(w, fakeResourceResponse{Resource: resource})
	case http.MethodDelete:
		f.mu.Lock()
		delete(f.resources[kind], name)
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (f *fakeBindplane) writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	require.NoError(f.t, json.NewEncoder(w).Encode(v))
}
