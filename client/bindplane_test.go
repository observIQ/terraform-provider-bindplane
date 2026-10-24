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
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/observiq/terraform-provider-bindplane/internal/fakebindplane"
	"github.com/observiq/terraform-provider-bindplane/internal/model"
	"github.com/stretchr/testify/require"
)

// Resources shaped like the Bindplane server's responses, including
// fields the provider does not use.
const (
	seedConfiguration = `{
		"apiVersion": "bindplane.observiq.com/v2",
		"kind": "Configuration",
		"metadata": {
			"id": "cfg-id",
			"name": "my-config",
			"version": 2,
			"hash": "abc123",
			"dateModified": "2026-01-02T03:04:05Z",
			"labels": {"platform": "linux", "env": "test"}
		},
		"spec": {
			"contentType": "text/yaml",
			"measurementInterval": "1m",
			"topologyInterval": "",
			"parameters": [
				{"name": "telemetryPort", "value": 8888},
				{"name": "telemetryLevel", "value": "detailed"}
			],
			"sources": [{
				"id": "src-route",
				"name": "my-source:3",
				"processors": [{"name": "my-batch:1"}],
				"routes": {"logs": [{"id": "r1", "components": ["destinations/dest-route"]}]}
			}],
			"processors": [{
				"id": "pg-1",
				"processors": [{"name": "my-batch:1"}],
				"routes": {"metrics": [{"id": "r2", "components": ["destinations/dest-route"]}]}
			}],
			"connectors": [{
				"id": "conn-1",
				"name": "my-connector:1",
				"routes": {"traces": [{"id": "r3", "components": ["processors/pg-1"]}]}
			}],
			"destinations": [{"id": "dest-route", "name": "my-dest:2", "processors": [{"name": "my-filter:1"}]}],
			"extensions": [{"name": "my-pprof:1"}],
			"selector": {"matchLabels": {"configuration": "my-config"}},
			"rollout": {
				"type": "progressive",
				"parameters": [{"name": "stages", "value": [{"labels": {"env": "stage"}, "name": "stage"}]}]
			}
		},
		"status": {"rollout": {"status": 4}}
	}`

	seedFleet = `{
		"apiVersion": "bindplane.observiq.com/v1",
		"kind": "Fleet",
		"metadata": {
			"id": "my-fleet",
			"name": "my-fleet",
			"displayName": "My Fleet",
			"version": 1,
			"labels": {"agent-type": "observiq-otel-collector", "platform": "linux"}
		},
		"spec": {
			"configuration": "my-config",
			"selector": {"matchLabels": {"fleet": "my-fleet"}}
		},
		"status": {}
	}`
)

// seedComponent returns a source, destination, processor, extension, or
// connector shaped like the Bindplane server's responses.
func seedComponent(kind, name string) string {
	return `{
		"apiVersion": "bindplane.observiq.com/v1",
		"kind": "` + kind + `",
		"metadata": {"id": "` + name + `-id", "name": "` + name + `", "version": 3, "labels": {}},
		"spec": {
			"type": "custom:2",
			"parameters": [
				{"name": "collection_interval", "value": 60},
				{"name": "password", "value": "(sensitive)", "sensitive": true},
				{"name": "token", "value": "abc", "credential": true}
			],
			"processors": [{"name": "my-batch:1"}]
		},
		"status": {"latest": true}
	}`
}

func newFakeClient(t *testing.T, srv *fakebindplane.Server, cfg Config) *BindPlane {
	t.Helper()
	cfg.RemoteURL = srv.URL
	c, err := New(cfg)
	require.NoError(t, err)
	return c
}

func testSourceResource() *model.AnyResource {
	return &model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       model.KindSource,
			Metadata: model.Metadata{
				ID:   "src-id",
				Name: "my-source",
			},
		},
		Spec: map[string]any{
			"type": "host",
			"parameters": []model.Parameter{
				{Name: "interval", Value: 60},
				{Name: "password", Value: "secret", Sensitive: true},
			},
		},
	}
}

func testConfigurationResource(measurementInterval string) *model.AnyResource {
	return &model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       model.KindConfiguration,
			Metadata: model.Metadata{
				Name:   "my-config",
				Labels: model.LabelsFromValidatedMap(map[string]string{"platform": "linux"}),
			},
		},
		Spec: map[string]any{
			"contentType":         "text/yaml",
			"measurementInterval": measurementInterval,
		},
	}
}

func TestBindPlane_Apply_request(t *testing.T) {
	srv := fakebindplane.New(t)
	c := newFakeClient(t, srv, Config{})

	require.NoError(t, c.Apply(testSourceResource(), false))

	req := srv.LastRequest(http.MethodPost, "/v1/apply")
	require.NotNil(t, req)
	require.Equal(t, "acceptInvalid=false", req.Query.Encode())
	require.Equal(t, "application/json", req.Header.Get("Content-Type"))
	require.JSONEq(t, `{
		"resources": [{
			"apiVersion": "bindplane.observiq.com/v1",
			"kind": "Source",
			"metadata": {"id": "src-id", "name": "my-source", "labels": {}},
			"spec": {
				"type": "host",
				"parameters": [
					{"name": "interval", "value": 60},
					{"name": "password", "value": "secret", "sensitive": true}
				]
			}
		}]
	}`, string(req.Body))

	require.NotNil(t, srv.Resource("Source", "my-source"))
}

func TestBindPlane_Apply_status(t *testing.T) {
	cases := []struct {
		status    string
		expectErr bool
	}{
		{"created", false},
		{"configured", false},
		{"unchanged", false},
		{"invalid", true},
		{"error", true},
		{"in-use", true},
		{"forbidden", true},
	}

	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			srv := fakebindplane.New(t)
			c := newFakeClient(t, srv, Config{})

			srv.RespondOnce(http.MethodPost, "/v1/apply", fakebindplane.Response{
				Status: http.StatusAccepted,
				Body: `{"updates": [{
					"resource": {"apiVersion": "bindplane.observiq.com/v1", "kind": "Source", "metadata": {"id": "src-id", "name": "my-source"}, "spec": {}},
					"status": "` + tc.status + `",
					"reason": "the reason"
				}]}`,
			})

			err := c.Apply(testSourceResource(), false)
			if !tc.expectErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.ErrorContains(t, err, "my-source")
			require.ErrorContains(t, err, "status: "+tc.status)
			require.ErrorContains(t, err, "reason: the reason")
		})
	}
}

func TestBindPlane_Apply_rollout(t *testing.T) {
	const rolloutPath = "/v1/rollouts/my-config/start"

	t.Run("created configuration", func(t *testing.T) {
		srv := fakebindplane.New(t)
		c := newFakeClient(t, srv, Config{})

		require.NoError(t, c.Apply(testConfigurationResource("10s"), true))

		req := srv.LastRequest(http.MethodPost, rolloutPath)
		require.NotNil(t, req, "expected a rollout to be started")
		require.Equal(t, "application/json", req.Header.Get("Content-Type"))
		require.JSONEq(t, `{"options": null}`, string(req.Body))
	})

	t.Run("configured configuration", func(t *testing.T) {
		srv := fakebindplane.New(t)
		c := newFakeClient(t, srv, Config{})

		require.NoError(t, c.Apply(testConfigurationResource("10s"), false))
		require.Empty(t, srv.RequestsTo(http.MethodPost, rolloutPath))

		require.NoError(t, c.Apply(testConfigurationResource("1m"), true))
		require.Len(t, srv.RequestsTo(http.MethodPost, rolloutPath), 1)
	})

	t.Run("unchanged configuration", func(t *testing.T) {
		srv := fakebindplane.New(t)
		c := newFakeClient(t, srv, Config{})

		require.NoError(t, c.Apply(testConfigurationResource("10s"), false))
		// The first apply assigned an ID, so re-apply the stored
		// resource to get an unchanged status.
		stored := testConfigurationResource("10s")
		stored.Metadata.ID = srv.Resource("Configuration", "my-config")["metadata"].(map[string]any)["id"].(string)
		require.NoError(t, c.Apply(stored, true))
		require.Empty(t, srv.RequestsTo(http.MethodPost, rolloutPath))
	})

	t.Run("not a configuration", func(t *testing.T) {
		srv := fakebindplane.New(t)
		c := newFakeClient(t, srv, Config{})

		require.NoError(t, c.Apply(testSourceResource(), true))
		for _, r := range srv.Requests() {
			require.False(t, strings.HasPrefix(r.Path, "/v1/rollouts/"), "unexpected rollout request %s", r.Path)
		}
	})

	t.Run("rollout fails", func(t *testing.T) {
		srv := fakebindplane.New(t)
		c := newFakeClient(t, srv, Config{})

		srv.RespondOnce(http.MethodPost, rolloutPath, fakebindplane.Response{
			Status: http.StatusConflict,
			Body:   `{"errors": ["rollout already in progress"]}`,
		})
		err := c.Apply(testConfigurationResource("10s"), true)
		require.ErrorContains(t, err, "409 Conflict")
		require.ErrorContains(t, err, "rollout already in progress")
	})
}

func TestBindPlane_Apply_httpError(t *testing.T) {
	srv := fakebindplane.New(t)
	c := newFakeClient(t, srv, Config{})

	srv.RespondOnce(http.MethodPost, "/v1/apply", fakebindplane.Response{
		Status: http.StatusInternalServerError,
		Body:   `{"errors": ["boom"]}`,
	})

	err := c.Apply(testSourceResource(), false)
	require.ErrorContains(t, err, "500 Internal Server Error: boom")
}

func TestBindPlane_ApplyWithRetry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := fakebindplane.New(t)
		c := newFakeClient(t, srv, Config{})

		require.NoError(t, c.ApplyWithRetry(context.Background(), time.Minute, testSourceResource(), false))
		require.Len(t, srv.RequestsTo(http.MethodPost, "/v1/apply"), 1)
	})

	t.Run("non-retryable error", func(t *testing.T) {
		srv := fakebindplane.New(t)
		c := newFakeClient(t, srv, Config{})

		srv.RespondOnce(http.MethodPost, "/v1/apply", fakebindplane.Response{
			Status: http.StatusBadRequest,
			Body:   `{"errors": ["bad resource"]}`,
		})

		err := c.ApplyWithRetry(context.Background(), time.Minute, testSourceResource(), false)
		require.ErrorContains(t, err, "retries exhausted")
		require.ErrorContains(t, err, "bad resource")
		require.Len(t, srv.RequestsTo(http.MethodPost, "/v1/apply"), 1, "non-retryable errors should not be retried")
	})

	t.Run("connection refused", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := l.Addr().String()
		require.NoError(t, l.Close())

		c, err := New(Config{RemoteURL: "http://" + addr})
		require.NoError(t, err)

		err = c.ApplyWithRetry(context.Background(), 2*time.Second, testSourceResource(), false)
		require.ErrorContains(t, err, "retries exhausted")
		require.ErrorContains(t, err, "connection refused")
	})
}

func TestBindPlane_Configuration(t *testing.T) {
	srv := fakebindplane.New(t)
	c := newFakeClient(t, srv, Config{})
	srv.Seed(t, seedConfiguration)

	config, err := c.Configuration("my-config")
	require.NoError(t, err)
	require.NotNil(t, config)

	require.NotNil(t, srv.LastRequest(http.MethodGet, "/v1/configurations/my-config"))

	require.Equal(t, "cfg-id", config.ID())
	require.Equal(t, "my-config", config.Name())
	require.Equal(t, "bindplane.observiq.com/v2", config.APIVersion)
	require.Equal(t, model.KindConfiguration, config.Kind)
	require.Equal(t, map[string]string{"platform": "linux", "env": "test"}, config.Metadata.Labels.AsMap())

	spec := config.Spec
	require.Equal(t, "text/yaml", spec.ContentType)
	require.Equal(t, "1m", spec.MeasurementInterval)
	require.Equal(t, "my-config", spec.Selector.MatchLabels["configuration"])

	require.Len(t, spec.Parameters, 2)
	require.Equal(t, "telemetryPort", spec.Parameters[0].Name)
	require.Equal(t, float64(8888), spec.Parameters[0].Value)
	require.Equal(t, "telemetryLevel", spec.Parameters[1].Name)
	require.Equal(t, "detailed", spec.Parameters[1].Value)

	require.Len(t, spec.Sources, 1)
	require.Equal(t, "src-route", spec.Sources[0].ID)
	require.Equal(t, "my-source:3", spec.Sources[0].Name)
	require.Equal(t, "my-batch:1", spec.Sources[0].Processors[0].Name)
	require.NotNil(t, spec.Sources[0].Routes)
	require.Len(t, spec.Sources[0].Routes.Logs, 1)
	require.Equal(t, "r1", spec.Sources[0].Routes.Logs[0].ID)
	require.Equal(t, []model.ComponentPath{"destinations/dest-route"}, spec.Sources[0].Routes.Logs[0].Components)

	require.Len(t, spec.Processors, 1)
	require.Equal(t, "pg-1", spec.Processors[0].ID)
	require.Equal(t, "my-batch:1", spec.Processors[0].Processors[0].Name)
	require.Equal(t, "r2", spec.Processors[0].Routes.Metrics[0].ID)

	require.Len(t, spec.Connectors, 1)
	require.Equal(t, "conn-1", spec.Connectors[0].ID)
	require.Equal(t, "my-connector:1", spec.Connectors[0].Name)
	require.Equal(t, []model.ComponentPath{"processors/pg-1"}, spec.Connectors[0].Routes.Traces[0].Components)

	require.Len(t, spec.Destinations, 1)
	require.Equal(t, "dest-route", spec.Destinations[0].ID)
	require.Equal(t, "my-dest:2", spec.Destinations[0].Name)
	require.Equal(t, "my-filter:1", spec.Destinations[0].Processors[0].Name)
	require.Nil(t, spec.Destinations[0].Routes)

	require.Len(t, spec.Extensions, 1)
	require.Equal(t, "my-pprof:1", spec.Extensions[0].Name)

	require.Equal(t, "progressive", spec.Rollout.Type)
	require.Len(t, spec.Rollout.Parameters, 1)
	require.Equal(t, "stages", spec.Rollout.Parameters[0].Name)
	require.Equal(t, []any{map[string]any{"labels": map[string]any{"env": "stage"}, "name": "stage"}}, spec.Rollout.Parameters[0].Value)
}

func TestBindPlane_Fleet(t *testing.T) {
	srv := fakebindplane.New(t)
	c := newFakeClient(t, srv, Config{})
	srv.Seed(t, seedFleet)

	fleet, err := c.Fleet("my-fleet")
	require.NoError(t, err)
	require.NotNil(t, fleet)

	require.NotNil(t, srv.LastRequest(http.MethodGet, "/v1/resources/Fleet/my-fleet"))

	require.Equal(t, "my-fleet", fleet.ID())
	require.Equal(t, "my-fleet", fleet.Name())
	require.Equal(t, "My Fleet", fleet.Metadata.DisplayName)
	require.Equal(t, "observiq-otel-collector", fleet.GetLabels().Set["agent-type"])
	require.Equal(t, "linux", fleet.GetLabels().Set["platform"])
	require.Equal(t, "my-config", fleet.FleetSpec.Configuration)
}

func TestBindPlane_GenericResource(t *testing.T) {
	cases := []struct {
		kind model.Kind
		path string
	}{
		{model.KindSource, "/v1/sources/my-component"},
		{model.KindDestination, "/v1/destinations/my-component"},
		{model.KindProcessor, "/v1/processors/my-component"},
		{model.KindExtension, "/v1/extensions/my-component"},
		{model.KindConnector, "/v1/resources/Connector/my-component"},
	}

	for _, tc := range cases {
		t.Run(string(tc.kind), func(t *testing.T) {
			srv := fakebindplane.New(t)
			c := newFakeClient(t, srv, Config{})
			srv.Seed(t, seedComponent(string(tc.kind), "my-component"))

			g, err := c.GenericResource(tc.kind, "my-component")
			require.NoError(t, err)
			require.NotNil(t, g)

			req := srv.LastRequest(http.MethodGet, tc.path)
			require.NotNil(t, req)
			require.Empty(t, req.Header.Get("Content-Type"), "GET requests should not have a content type")

			require.Equal(t, "my-component-id", g.ID)
			require.Equal(t, "my-component", g.Name)
			require.Equal(t, "custom:2", g.Spec.Type)

			require.Len(t, g.Spec.Parameters, 3)
			require.Equal(t, "collection_interval", g.Spec.Parameters[0].Name)
			require.Equal(t, float64(60), g.Spec.Parameters[0].Value)
			require.False(t, g.Spec.Parameters[0].Sensitive)
			require.Equal(t, "password", g.Spec.Parameters[1].Name)
			require.Equal(t, "(sensitive)", g.Spec.Parameters[1].Value)
			require.True(t, g.Spec.Parameters[1].Sensitive)
			require.Equal(t, "token", g.Spec.Parameters[2].Name)
			require.True(t, g.Spec.Parameters[2].Credential)

			require.Len(t, g.Spec.Processors, 1)
			require.Equal(t, "my-batch:1", g.Spec.Processors[0].Name)

			// Missing resources are not an error.
			g, err = c.GenericResource(tc.kind, "missing")
			require.NoError(t, err)
			require.Nil(t, g)
		})
	}

	t.Run("unsupported kind", func(t *testing.T) {
		srv := fakebindplane.New(t)
		c := newFakeClient(t, srv, Config{})

		for _, kind := range []model.Kind{model.KindAgent, model.KindConfiguration, model.KindFleet} {
			_, err := c.GenericResource(kind, "name")
			require.ErrorContains(t, err, "does not support bindplane kind")
		}
		require.Empty(t, srv.Requests())
	})
}

func TestBindPlane_getters(t *testing.T) {
	cases := []struct {
		name string
		path string
		get  func(c *BindPlane, name string) (bool, error)
	}{
		{"configuration", "/v1/configurations/", func(c *BindPlane, n string) (bool, error) { r, err := c.Configuration(n); return r != nil, err }},
		{"source", "/v1/sources/", func(c *BindPlane, n string) (bool, error) { r, err := c.Source(n); return r != nil, err }},
		{"destination", "/v1/destinations/", func(c *BindPlane, n string) (bool, error) { r, err := c.Destination(n); return r != nil, err }},
		{"processor", "/v1/processors/", func(c *BindPlane, n string) (bool, error) { r, err := c.Processor(n); return r != nil, err }},
		{"extension", "/v1/extensions/", func(c *BindPlane, n string) (bool, error) { r, err := c.Extension(n); return r != nil, err }},
		{"connector", "/v1/resources/Connector/", func(c *BindPlane, n string) (bool, error) { r, err := c.Connector(n); return r != nil, err }},
		{"fleet", "/v1/resources/Fleet/", func(c *BindPlane, n string) (bool, error) { r, err := c.Fleet(n); return r != nil, err }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("not found", func(t *testing.T) {
				srv := fakebindplane.New(t)
				c := newFakeClient(t, srv, Config{})

				found, err := tc.get(c, "missing")
				require.NoError(t, err, "a missing resource is not an error")
				require.False(t, found)
				require.NotNil(t, srv.LastRequest(http.MethodGet, tc.path+"missing"))
			})

			t.Run("server error", func(t *testing.T) {
				srv := fakebindplane.New(t)
				c := newFakeClient(t, srv, Config{})

				srv.RespondOnce(http.MethodGet, tc.path+"name", fakebindplane.Response{
					Status: http.StatusInternalServerError,
					Body:   `{"errors": ["database unavailable"]}`,
				})
				_, err := tc.get(c, "name")
				require.ErrorContains(t, err, "500 Internal Server Error")
				require.ErrorContains(t, err, "database unavailable")
			})

			t.Run("unauthorized", func(t *testing.T) {
				srv := fakebindplane.New(t, fakebindplane.WithAuth(fakebindplane.Auth{APIKey: "right"}))
				c := newFakeClient(t, srv, Config{APIKey: "wrong"})

				_, err := tc.get(c, "name")
				require.ErrorContains(t, err, "401 Unauthorized")
			})

			t.Run("policies not accepted", func(t *testing.T) {
				srv := fakebindplane.New(t)
				c := newFakeClient(t, srv, Config{})

				srv.RespondOnce(http.MethodGet, tc.path+"name", fakebindplane.Response{Status: 453})
				_, err := tc.get(c, "name")
				require.ErrorContains(t, err, "policies not accepted")
			})

			t.Run("invalid response", func(t *testing.T) {
				srv := fakebindplane.New(t)
				c := newFakeClient(t, srv, Config{})

				srv.RespondOnce(http.MethodGet, tc.path+"name", fakebindplane.Response{
					Status: http.StatusOK,
					Body:   `{"resource": {"metadata": `,
				})
				_, err := tc.get(c, "name")
				require.Error(t, err)
			})

			// Servers which do not support an endpoint serve the UI.
			t.Run("html response", func(t *testing.T) {
				srv := fakebindplane.New(t)
				c := newFakeClient(t, srv, Config{})

				srv.RespondOnce(http.MethodGet, tc.path+"name", fakebindplane.Response{
					Status:      http.StatusOK,
					ContentType: "text/html; charset=utf-8",
					Body:        `<html></html>`,
				})
				_, err := tc.get(c, "name")
				require.ErrorContains(t, err, "endpoint not found")
			})
		})
	}
}

func TestBindPlane_Delete(t *testing.T) {
	cases := []struct {
		kind model.Kind
		seed string
		path string
	}{
		{model.KindConfiguration, seedConfiguration, "/v1/configurations/my-config"},
		{model.KindFleet, seedFleet, "/v1/resources/Fleet/my-fleet"},
		{model.KindSource, seedComponent("Source", "my-component"), "/v1/sources/my-component"},
		{model.KindDestination, seedComponent("Destination", "my-component"), "/v1/destinations/my-component"},
		{model.KindProcessor, seedComponent("Processor", "my-component"), "/v1/processors/my-component"},
		{model.KindExtension, seedComponent("Extension", "my-component"), "/v1/extensions/my-component"},
		{model.KindConnector, seedComponent("Connector", "my-component"), "/v1/resources/Connector/my-component"},
	}

	for _, tc := range cases {
		t.Run(string(tc.kind), func(t *testing.T) {
			srv := fakebindplane.New(t)
			c := newFakeClient(t, srv, Config{})
			srv.Seed(t, tc.seed)

			name := tc.path[strings.LastIndex(tc.path, "/")+1:]
			require.NotNil(t, srv.Resource(string(tc.kind), name))

			require.NoError(t, c.Delete(tc.kind, name))

			req := srv.LastRequest(http.MethodDelete, tc.path)
			require.NotNil(t, req)
			require.Equal(t, "force=false", req.Query.Encode())
			require.Nil(t, srv.Resource(string(tc.kind), name))

			t.Run("not found", func(t *testing.T) {
				err := c.Delete(tc.kind, name)
				require.Error(t, err)
				require.ErrorContains(t, err, name)
			})

			t.Run("in use", func(t *testing.T) {
				srv.RespondOnce(http.MethodDelete, tc.path, fakebindplane.Response{
					Status: http.StatusConflict,
					Body:   `{"errors": ["dependent resources: Configuration test"]}`,
				})
				err := c.Delete(tc.kind, name)
				require.ErrorContains(t, err, "dependent resources: Configuration test")
			})
		})
	}

	t.Run("unsupported kind", func(t *testing.T) {
		srv := fakebindplane.New(t)
		c := newFakeClient(t, srv, Config{})

		err := c.Delete(model.KindAgent, "agent")
		require.ErrorContains(t, err, "does not support bindplane kind")
		require.Empty(t, srv.Requests())
	})
}

func TestNew_auth(t *testing.T) {
	cases := []struct {
		name          string
		config        Config
		expectAPIKey  string
		expectUser    string
		expectPass    string
		expectAccount string
	}{
		{
			name: "no credentials",
		},
		{
			name:         "api key",
			config:       Config{APIKey: "legacy-key"},
			expectAPIKey: "legacy-key",
		},
		{
			name:         "api key takes precedence over basic auth",
			config:       Config{APIKey: "legacy-key", Username: "user", Password: "pass"},
			expectAPIKey: "legacy-key",
		},
		{
			name:          "scoped api key with account",
			config:        Config{APIKey: "bps_abc.def", AccountID: "01ACCOUNT"},
			expectAPIKey:  "bps_abc.def",
			expectAccount: "01ACCOUNT",
		},
		{
			name:       "basic auth",
			config:     Config{Username: "user", Password: "pass"},
			expectUser: "user",
			expectPass: "pass",
		},
		{
			name:   "username without password",
			config: Config{Username: "user"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := fakebindplane.New(t)
			c := newFakeClient(t, srv, tc.config)

			// Exercise a GET, POST, and DELETE request
			_, err := c.Configuration("missing")
			require.NoError(t, err)
			require.NoError(t, c.Apply(testSourceResource(), false))
			require.NoError(t, c.DeleteSource("my-source"))

			requests := srv.Requests()
			require.Len(t, requests, 3)
			for _, req := range requests {
				require.Equal(t, tc.expectAPIKey, req.Header.Get("X-Bindplane-Api-Key"), req.Method)
				require.Equal(t, tc.expectAccount, req.Header.Get("X-Bindplane-Account-ID"), req.Method)

				r := &http.Request{Header: req.Header}
				user, pass, ok := r.BasicAuth()
				require.Equal(t, tc.expectUser != "", ok, req.Method)
				require.Equal(t, tc.expectUser, user, req.Method)
				require.Equal(t, tc.expectPass, pass, req.Method)
			}
		})
	}
}

func TestNew_tls(t *testing.T) {
	tlsServer := func(t *testing.T, cfg *tls.Config) *fakebindplane.Server {
		return fakebindplane.New(t, fakebindplane.WithTLS(cfg))
	}

	mtlsServer := func(t *testing.T, cert fakebindplane.ClientCertificate) *fakebindplane.Server {
		return tlsServer(t, &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  cert.CAPool,
		})
	}

	// Each case makes a request and expects it to either succeed,
	// or fail with an error containing expectErr.
	cases := []struct {
		name      string
		setup     func(t *testing.T) (*fakebindplane.Server, Config)
		expectErr string
	}{
		{
			name: "certificate authority",
			setup: func(t *testing.T) (*fakebindplane.Server, Config) {
				srv := tlsServer(t, nil)
				return srv, Config{TLSCertificateAuthority: srv.CAFile}
			},
		},
		{
			name: "untrusted certificate",
			setup: func(t *testing.T) (*fakebindplane.Server, Config) {
				return tlsServer(t, nil), Config{}
			},
			expectErr: "certificate",
		},
		{
			name: "skip verify",
			setup: func(t *testing.T) (*fakebindplane.Server, Config) {
				return tlsServer(t, nil), Config{TLSSkipVerify: true}
			},
		},
		{
			name: "mutual tls",
			setup: func(t *testing.T) (*fakebindplane.Server, Config) {
				cert := fakebindplane.NewClientCertificate(t)
				srv := mtlsServer(t, cert)
				return srv, Config{
					TLSCertificateAuthority: srv.CAFile,
					TLSCertificate:          cert.CertFile,
					TLSPrivateKey:           cert.KeyFile,
				}
			},
		},
		{
			name: "mutual tls without client certificate",
			setup: func(t *testing.T) (*fakebindplane.Server, Config) {
				srv := mtlsServer(t, fakebindplane.NewClientCertificate(t))
				return srv, Config{TLSCertificateAuthority: srv.CAFile}
			},
			expectErr: "certificate required",
		},
		{
			// Bindplane's client requires TLS 1.3 or newer
			name: "tls 1.2 server",
			setup: func(t *testing.T) (*fakebindplane.Server, Config) {
				srv := tlsServer(t, &tls.Config{MaxVersion: tls.VersionTLS12})
				return srv, Config{TLSCertificateAuthority: srv.CAFile}
			},
			expectErr: "protocol version",
		},
		{
			name: "https client and http server",
			setup: func(t *testing.T) (*fakebindplane.Server, Config) {
				srv := fakebindplane.New(t)
				return srv, Config{RemoteURL: strings.Replace(srv.URL, "http://", "https://", 1)}
			},
			expectErr: "server gave HTTP response to HTTPS client",
		},
		{
			name: "http client and https server",
			setup: func(t *testing.T) (*fakebindplane.Server, Config) {
				srv := tlsServer(t, nil)
				return srv, Config{RemoteURL: strings.Replace(srv.URL, "https://", "http://", 1)}
			},
			expectErr: "400 Bad Request",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, cfg := tc.setup(t)
			if cfg.RemoteURL == "" {
				cfg.RemoteURL = srv.URL
			}
			c, err := New(cfg)
			require.NoError(t, err)

			config, err := c.Configuration("missing")
			if tc.expectErr != "" {
				require.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			require.Nil(t, config)
			require.Len(t, srv.Requests(), 1)
		})
	}
}

func TestNew_invalidTLSFiles(t *testing.T) {
	notPEM := filepath.Join(t.TempDir(), "not-pem")
	require.NoError(t, os.WriteFile(notPEM, []byte("not a certificate"), 0o600))

	srv := httptest.NewTLSServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)

	cases := []struct {
		name   string
		config Config
	}{
		{"missing certificate authority", Config{TLSCertificateAuthority: filepath.Join(t.TempDir(), "missing.crt")}},
		{"invalid certificate authority", Config{TLSCertificateAuthority: notPEM}},
		{"missing client certificate", Config{TLSCertificate: filepath.Join(t.TempDir(), "missing.crt"), TLSPrivateKey: filepath.Join(t.TempDir(), "missing.key")}},
		{"invalid client certificate", Config{TLSCertificate: notPEM, TLSPrivateKey: notPEM}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.config.RemoteURL = srv.URL
			_, err := New(tc.config)
			require.Error(t, err)
		})
	}
}
