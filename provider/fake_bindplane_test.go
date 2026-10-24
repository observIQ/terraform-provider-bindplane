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
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/stretchr/testify/require"
)

// fakeServerErrorName is a resource name the fake Bindplane server always
// answers with an internal server error.
const fakeServerErrorName = "server-error"

// fakeBindplane is an in-memory Bindplane API serving the endpoints used by
// the configuration and processor bundle resources. Applied resources are
// recorded and stored, so that a following read returns them.
type fakeBindplane struct {
	t *testing.T

	mu             sync.Mutex
	configurations map[string]*model.Configuration
	processors     map[string]*model.Processor
	applied        []*model.AnyResource
	rollouts       []string
}

// newFakeBindplane starts a fake Bindplane server and returns it with a
// client configured to use it, as the provider passes it to resources.
func newFakeBindplane(t *testing.T) (*fakeBindplane, *client.BindPlane) {
	f := &fakeBindplane{
		t:              t,
		configurations: map[string]*model.Configuration{},
		processors:     map[string]*model.Processor{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/configurations/{name}", f.getConfiguration)
	mux.HandleFunc("GET /v1/processors/{name}", f.getProcessor)
	mux.HandleFunc("POST /v1/apply", f.apply)
	mux.HandleFunc("POST /v1/rollouts/{name}/start", f.startRollout)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	d := schema.TestResourceDataRaw(t, Provider().Schema, map[string]any{
		"remote_url": srv.URL,
		"username":   "user",
		"password":   "pass",
	})
	meta, diags := providerConfigure(d, nil)
	require.False(t, diags.HasError(), "configure provider: %v", diags)

	return f, meta.(*client.BindPlane)
}

func (f *fakeBindplane) getConfiguration(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if r.PathValue("name") == fakeServerErrorName {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	c, ok := f.configurations[r.PathValue("name")]
	if !ok {
		http.NotFound(w, r)
		return
	}
	f.writeJSON(w, model.ConfigurationResponse{Configuration: c})
}

func (f *fakeBindplane) getProcessor(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if r.PathValue("name") == fakeServerErrorName {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	p, ok := f.processors[r.PathValue("name")]
	if !ok {
		http.NotFound(w, r)
		return
	}
	f.writeJSON(w, model.ProcessorResponse{Processor: p})
}

func (f *fakeBindplane) apply(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	payload := model.ApplyPayload{}
	require.NoError(f.t, json.NewDecoder(r.Body).Decode(&payload))

	resp := model.ApplyResponseClientSide{}
	for _, res := range payload.Resources {
		f.applied = append(f.applied, res)

		raw, err := json.Marshal(res)
		require.NoError(f.t, err)

		switch res.Kind {
		case model.KindConfiguration:
			c := &model.Configuration{}
			require.NoError(f.t, json.Unmarshal(raw, c))
			if existing, ok := f.configurations[c.Name()]; ok {
				c.Metadata.ID = existing.Metadata.ID
			} else if c.Metadata.ID == "" {
				c.Metadata.ID = "configuration-" + c.Name()
			}
			f.configurations[c.Name()] = c
		case model.KindProcessor:
			p := &model.Processor{}
			require.NoError(f.t, json.Unmarshal(raw, p))
			f.processors[p.Name()] = p
		default:
			f.t.Errorf("fake bindplane: unsupported kind %s", res.Kind)
		}

		resp.Updates = append(resp.Updates, &model.AnyResourceStatus{
			Resource: *res,
			Status:   model.StatusCreated,
		})
	}
	f.writeJSON(w, resp)
}

func (f *fakeBindplane) startRollout(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	name := r.PathValue("name")
	f.rollouts = append(f.rollouts, name)
	f.writeJSON(w, model.ConfigurationResponse{Configuration: f.configurations[name]})
}

func (f *fakeBindplane) writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	require.NoError(f.t, json.NewEncoder(w).Encode(v))
}

// lastApplied returns the most recent resource sent to /v1/apply.
func (f *fakeBindplane) lastApplied() *model.AnyResource {
	f.mu.Lock()
	defer f.mu.Unlock()

	require.NotEmpty(f.t, f.applied, "no resources were applied")
	return f.applied[len(f.applied)-1]
}
