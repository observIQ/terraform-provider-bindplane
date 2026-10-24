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

// Package fakebindplane provides an in-memory fake of the Bindplane REST API
// for use in tests. It implements only the endpoints used by the provider, using
// the same wire format as the Bindplane server. Resources are stored as raw JSON
// objects so that the fake is independent of any Go model package.
package fakebindplane

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

const (
	headerAPIKey    = "X-Bindplane-Api-Key"    // #nosec G101 this is a header name, not a credential
	headerAccountID = "X-Bindplane-Account-ID" // #nosec G101 this is a header name, not a credential
)

// collections maps REST collection names to their resource kind.
var collections = map[string]string{
	"configurations": "Configuration",
	"sources":        "Source",
	"destinations":   "Destination",
	"processors":     "Processor",
	"extensions":     "Extension",
}

// Auth is the set of credentials the fake server requires. Empty
// fields are not checked.
type Auth struct {
	APIKey    string
	Username  string
	Password  string
	AccountID string
}

// Request is a request received by the fake server.
type Request struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   []byte
}

// Response is a canned response returned by the fake server.
type Response struct {
	Status      int
	ContentType string
	Body        string
}

// applyRequest is the body of POST /v1/apply
type applyRequest struct {
	Resources []map[string]any `json:"resources"`
}

// applyResponse is the response to POST /v1/apply
type applyResponse struct {
	Updates []resourceStatus `json:"updates"`
}

// resourceStatus is the result of applying a resource
type resourceStatus struct {
	Resource map[string]any `json:"resource"`
	Status   string         `json:"status"`
	Reason   string         `json:"reason"`
}

// rolloutResponse is the response to POST /v1/rollouts/{name}/start
type rolloutResponse struct {
	Configuration map[string]any `json:"configuration"`
}

// resourceResponse is the response to GET /v1/resources/{kind}/{name}
type resourceResponse struct {
	Resource map[string]any `json:"resource"`
}

// errorResponse is the body of an unsuccessful response
type errorResponse struct {
	Errors []string `json:"errors"`
}

type override struct {
	method   string
	path     string
	response Response
}

// Option configures a Server.
type Option func(*Server)

// WithAuth requires requests to carry the given credentials.
func WithAuth(a Auth) Option {
	return func(s *Server) {
		s.auth = a
	}
}

// WithTLS serves over TLS using the given base configuration, which
// may be nil. The server certificate is the httptest certificate, which
// is written to CAFile so that clients can trust it.
func WithTLS(cfg *tls.Config) Option {
	return func(s *Server) {
		s.tls = true
		s.tlsConfig = cfg
	}
}

// Server is a fake Bindplane server.
type Server struct {
	// URL is the base URL of the server, without the /v1 prefix.
	URL string

	// CAFile is the path to a PEM file containing the server's certificate
	// when the server is using TLS.
	CAFile string

	auth      Auth
	tls       bool
	tlsConfig *tls.Config

	mu        sync.Mutex
	resources map[string]map[string]map[string]any
	requests  []Request
	overrides []override
	nextID    int
}

// New starts a fake Bindplane server that is closed when the test finishes.
func New(t testing.TB, opts ...Option) *Server {
	t.Helper()

	s := &Server{
		resources: map[string]map[string]map[string]any{},
	}
	for _, opt := range opts {
		opt(s)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/apply", s.handleApply)
	mux.HandleFunc("POST /v1/rollouts/{name}/start", s.handleRollout)
	mux.HandleFunc("GET /v1/resources/{kind}/{name}", s.handleGetResource)
	mux.HandleFunc("DELETE /v1/resources/{kind}/{name}", s.handleDeleteResource)
	mux.HandleFunc("GET /v1/{collection}/{name}", s.handleGet)
	mux.HandleFunc("DELETE /v1/{collection}/{name}", s.handleDelete)

	srv := httptest.NewUnstartedServer(s.middleware(mux))
	if s.tls {
		srv.TLS = s.tlsConfig
		srv.StartTLS()
		s.CAFile = writePEM(t, "server-ca.crt", "CERTIFICATE", srv.Certificate().Raw)
	} else {
		srv.Start()
	}
	t.Cleanup(srv.Close)

	s.URL = srv.URL
	return s
}

// Seed stores a resource, given as JSON, as if it had been created on
// the server. The resource's kind and name are read from the JSON.
func (s *Server) Seed(t testing.TB, resource string) {
	t.Helper()

	r := map[string]any{}
	if err := json.Unmarshal([]byte(resource), &r); err != nil {
		t.Fatalf("seed: unmarshal resource: %v", err)
	}

	kind, name := kindAndName(r)
	if kind == "" || name == "" {
		t.Fatalf("seed: resource must have a kind and metadata.name: %s", resource)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.store(kind, name, r)
}

// Resource returns the stored resource with the given kind and name,
// or nil if it does not exist.
func (s *Server) Resource(kind, name string) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.resources[kind][name]
}

// RespondOnce causes the next request matching method and path to receive
// the given response instead of being handled normally.
func (s *Server) RespondOnce(method, path string, r Response) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.overrides = append(s.overrides, override{method: method, path: path, response: r})
}

// Requests returns all requests received by the server.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Request(nil), s.requests...)
}

// RequestsTo returns the requests received for the given method and path.
func (s *Server) RequestsTo(method, path string) []Request {
	out := []Request{}
	for _, r := range s.Requests() {
		if r.Method == method && r.Path == path {
			out = append(out, r)
		}
	}
	return out
}

// LastRequest returns the most recent request for the given method and
// path, or nil if there was none.
func (s *Server) LastRequest(method, path string) *Request {
	requests := s.RequestsTo(method, path)
	if len(requests) == 0 {
		return nil
	}
	return &requests[len(requests)-1]
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeErrors(w, http.StatusBadRequest, err.Error())
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		s.mu.Lock()
		s.requests = append(s.requests, Request{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.Query(),
			Header: r.Header.Clone(),
			Body:   body,
		})
		s.mu.Unlock()

		if !s.authorized(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if o, ok := s.popOverride(r.Method, r.URL.Path); ok {
			contentType := o.ContentType
			if contentType == "" {
				contentType = "application/json"
			}
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(o.Status)
			_, _ = io.WriteString(w, o.Body)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) authorized(r *http.Request) bool {
	if s.auth.APIKey != "" && r.Header.Get(headerAPIKey) != s.auth.APIKey {
		return false
	}
	if s.auth.Username != "" || s.auth.Password != "" {
		u, p, ok := r.BasicAuth()
		if !ok || u != s.auth.Username || p != s.auth.Password {
			return false
		}
	}
	if s.auth.AccountID != "" && r.Header.Get(headerAccountID) != s.auth.AccountID {
		return false
	}
	return true
}

func (s *Server) popOverride(method, path string) (Response, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, o := range s.overrides {
		if o.method == method && o.path == path {
			s.overrides = append(s.overrides[:i], s.overrides[i+1:]...)
			return o.response, true
		}
	}
	return Response{}, false
}

func (s *Server) handleApply(w http.ResponseWriter, r *http.Request) {
	var payload applyRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeErrors(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := applyResponse{Updates: []resourceStatus{}}

	s.mu.Lock()
	for _, resource := range payload.Resources {
		kind, name := kindAndName(resource)
		metadata, _ := resource["metadata"].(map[string]any)
		if kind == "" || name == "" || metadata == nil {
			resp.Updates = append(resp.Updates, resourceStatus{Resource: resource, Status: "invalid", Reason: "kind and metadata.name are required"})
			continue
		}

		status := "created"
		existing := s.resources[kind][name]
		if existing != nil {
			// Bindplane keeps the existing ID when one is not provided.
			if id, _ := metadata["id"].(string); id == "" {
				metadata["id"] = existing["metadata"].(map[string]any)["id"]
			}
			status = "configured"
			if reflect.DeepEqual(existing, resource) {
				status = "unchanged"
			}
		} else if id, _ := metadata["id"].(string); id == "" {
			s.nextID++
			metadata["id"] = fmt.Sprintf("fake-%d", s.nextID)
		}

		s.store(kind, name, resource)
		resp.Updates = append(resp.Updates, resourceStatus{Resource: resource, Status: status})
	}
	s.mu.Unlock()

	writeJSON(w, http.StatusAccepted, resp)
}

func (s *Server) handleRollout(w http.ResponseWriter, r *http.Request) {
	c := s.Resource("Configuration", r.PathValue("name"))
	if c == nil {
		writeErrors(w, http.StatusNotFound, "resource not found")
		return
	}
	writeJSON(w, http.StatusAccepted, rolloutResponse{Configuration: c})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	kind, ok := collections[r.PathValue("collection")]
	if !ok {
		writeErrors(w, http.StatusNotFound, "unknown collection")
		return
	}

	resource := s.Resource(kind, r.PathValue("name"))
	if resource == nil {
		writeErrors(w, http.StatusNotFound, "resource not found")
		return
	}
	// The resource is returned under its kind, e.g. {"source": {...}}
	writeJSON(w, http.StatusOK, map[string]any{strings.ToLower(kind): resource})
}

func (s *Server) handleGetResource(w http.ResponseWriter, r *http.Request) {
	resource := s.Resource(r.PathValue("kind"), r.PathValue("name"))
	if resource == nil {
		writeErrors(w, http.StatusNotFound, "resource not found")
		return
	}
	writeJSON(w, http.StatusOK, resourceResponse{Resource: resource})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	kind, ok := collections[r.PathValue("collection")]
	if !ok {
		writeErrors(w, http.StatusNotFound, "unknown collection")
		return
	}
	s.delete(w, kind, r.PathValue("name"))
}

func (s *Server) handleDeleteResource(w http.ResponseWriter, r *http.Request) {
	s.delete(w, r.PathValue("kind"), r.PathValue("name"))
}

func (s *Server) delete(w http.ResponseWriter, kind, name string) {
	s.mu.Lock()
	_, ok := s.resources[kind][name]
	delete(s.resources[kind], name)
	s.mu.Unlock()

	if !ok {
		writeErrors(w, http.StatusNotFound, "resource not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// store must be called with s.mu held.
func (s *Server) store(kind, name string, resource map[string]any) {
	if s.resources[kind] == nil {
		s.resources[kind] = map[string]map[string]any{}
	}
	s.resources[kind][name] = resource
}

func kindAndName(resource map[string]any) (string, string) {
	kind, _ := resource["kind"].(string)
	metadata, _ := resource["metadata"].(map[string]any)
	name, _ := metadata["name"].(string)
	return kind, name
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErrors(w http.ResponseWriter, status int, errs ...string) {
	writeJSON(w, status, errorResponse{Errors: errs})
}

func writePEM(t testing.TB, name, blockType string, der []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	data := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}
