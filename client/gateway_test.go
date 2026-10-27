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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/observiq/terraform-provider-bindplane/internal/model"
	"github.com/stretchr/testify/require"
)

// gatewayServer serves GET and DELETE /v1/resources/:kind/:name for a
// single resource, and 404 for anything else.
type gatewayServer struct {
	path     string
	body     string
	status   int
	requests []string
}

func (s *gatewayServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.requests = append(s.requests, r.Method+" "+r.URL.Path)
	if r.URL.Path != s.path {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodDelete {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s.status)
	_, _ = w.Write([]byte(s.body))
}

func newGatewayTestClient(t *testing.T, s *gatewayServer) *BindPlane {
	server := httptest.NewServer(s)
	t.Cleanup(server.Close)
	i, err := newTestConfig(server.URL, "u", "p", "", "", "")
	require.NoError(t, err)
	return i
}

func TestGateway(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		s := &gatewayServer{
			path:   "/v1/resources/Gateway/my-gateway",
			status: http.StatusOK,
			body: `{"resource": {"apiVersion": "bindplane.observiq.com/v1", "kind": "Gateway",
				"metadata": {"id": "01ABC", "name": "my-gateway", "displayName": "My Gateway"},
				"spec": {"configuration": "gw-config", "loadBalancerHostname": "lb.example.com"}}}`,
		}
		i := newGatewayTestClient(t, s)

		g, err := i.Gateway("my-gateway")
		require.NoError(t, err)
		require.NotNil(t, g)
		require.Equal(t, "01ABC", g.ID())
		require.Equal(t, "my-gateway", g.Name())
		require.Equal(t, "My Gateway", g.Metadata.DisplayName)
		require.Equal(t, "gw-config", g.GatewaySpec.Configuration)
		require.Equal(t, "lb.example.com", g.GatewaySpec.LoadBalancerHostname)
		require.Equal(t, model.GatewayModeManual, g.Mode())
	})

	t.Run("not found", func(t *testing.T) {
		s := &gatewayServer{path: "/v1/resources/Gateway/other"}
		i := newGatewayTestClient(t, s)

		g, err := i.Gateway("my-gateway")
		require.NoError(t, err)
		require.Nil(t, g)
	})

	t.Run("server error", func(t *testing.T) {
		s := &gatewayServer{
			path:   "/v1/resources/Gateway/my-gateway",
			status: http.StatusInternalServerError,
		}
		i := newGatewayTestClient(t, s)

		g, err := i.Gateway("my-gateway")
		require.ErrorContains(t, err, "failed to get gateway with name my-gateway")
		require.Nil(t, g)
	})

	t.Run("unexpected resource type", func(t *testing.T) {
		s := &gatewayServer{
			path:   "/v1/resources/Gateway/my-gateway",
			status: http.StatusOK,
			body: `{"resource": {"apiVersion": "bindplane.observiq.com/v1", "kind": "Fleet",
				"metadata": {"id": "01ABC", "name": "my-gateway"}, "spec": {}}}`,
		}
		i := newGatewayTestClient(t, s)

		g, err := i.Gateway("my-gateway")
		require.ErrorContains(t, err, "expected gateway, got Fleet")
		require.Nil(t, g)
	})
}

func TestDeleteGateway(t *testing.T) {
	s := &gatewayServer{path: "/v1/resources/Gateway/my-gateway"}
	i := newGatewayTestClient(t, s)

	require.NoError(t, i.DeleteGateway("my-gateway"))
	require.NoError(t, i.Delete(model.KindGateway, "my-gateway"))
	require.Equal(t, []string{
		"DELETE /v1/resources/Gateway/my-gateway",
		"DELETE /v1/resources/Gateway/my-gateway",
	}, s.requests)

	err := i.DeleteGateway("missing")
	require.ErrorContains(t, err, "error while deleting gateway with name missing")
	require.Error(t, i.Delete(model.KindGateway, "missing"))
}

func TestGatewayUnsupportedError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		hint bool
	}{
		{"apply", errors.New("unable to apply resources, got 400 Bad Request: unknown resource kind: Gateway"), true},
		{"get", errors.New("got 400 Bad Request: kind: Gateway cannot be requested using the resource endpoint"), true},
		{"other kind", errors.New("unknown resource kind: Widget"), false},
		{"other error", errors.New("got 500 Internal Server Error"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := GatewayUnsupportedError(tc.err)
			require.ErrorIs(t, err, tc.err)
			if tc.hint {
				require.ErrorContains(t, err, "gateways require Bindplane v1.103.0 or newer")
				return
			}
			require.Equal(t, tc.err, err)
		})
	}
}
