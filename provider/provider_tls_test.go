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
	"crypto/tls"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/observiq/bindplane-op-enterprise/config"
	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/stretchr/testify/require"
)

// invalidTLSMinVersions are tls_min_version values that must be rejected.
var invalidTLSMinVersions = []string{"1.0", "1.1", "tls1.2", "1.4", "TLS1.3", " 1.2"}

func TestProvider_tlsMinVersion_schema(t *testing.T) {
	s := Configure().Schema["tls_min_version"]
	require.NotNil(t, s)
	require.Equal(t, schema.TypeString, s.Type)
	require.True(t, s.Optional)
	require.False(t, s.Required)
	require.Nil(t, s.Default, "Default and DefaultFunc cannot both be set")
	require.NotNil(t, s.DefaultFunc)
	require.NotNil(t, s.ValidateFunc)
	require.Contains(t, s.Description, "1.2")
	require.Contains(t, s.Description, "TLS 1.3")
}

func TestProvider_tlsMinVersion_default(t *testing.T) {
	t.Setenv(envTLSMinVersion, "")

	v, err := Configure().Schema["tls_min_version"].DefaultFunc()
	require.NoError(t, err)
	require.Equal(t, "1.3", v)
}

func TestProvider_tlsMinVersion_env(t *testing.T) {
	t.Setenv(envTLSMinVersion, "1.2")

	v, err := Configure().Schema["tls_min_version"].DefaultFunc()
	require.NoError(t, err)
	require.Equal(t, "1.2", v)
}

func TestProvider_tlsMinVersion_validate(t *testing.T) {
	validate := func(raw map[string]any) error {
		diags := Provider().Validate(terraform.NewResourceConfigRaw(raw))
		if diags.HasError() {
			return errorFromDiags(diags)
		}
		return nil
	}

	t.Run("unset", func(t *testing.T) {
		t.Setenv(envTLSMinVersion, "")
		require.NoError(t, validate(map[string]any{}))
	})

	for _, v := range []string{"1.2", "1.3"} {
		t.Run("attribute valid "+v, func(t *testing.T) {
			t.Setenv(envTLSMinVersion, "")
			require.NoError(t, validate(map[string]any{"tls_min_version": v}))
		})

		t.Run("env valid "+v, func(t *testing.T) {
			t.Setenv(envTLSMinVersion, v)
			require.NoError(t, validate(map[string]any{}))
		})
	}

	for _, v := range invalidTLSMinVersions {
		t.Run("attribute invalid "+v, func(t *testing.T) {
			t.Setenv(envTLSMinVersion, "")
			err := validate(map[string]any{"tls_min_version": v})
			require.Error(t, err)
			require.Contains(t, err.Error(), `"1.2"`)
			require.Contains(t, err.Error(), `"1.3"`)
		})

		t.Run("env invalid "+v, func(t *testing.T) {
			t.Setenv(envTLSMinVersion, v)
			err := validate(map[string]any{})
			require.Error(t, err)
			require.Contains(t, err.Error(), `"1.2"`)
			require.Contains(t, err.Error(), `"1.3"`)
		})
	}

	t.Run("attribute overrides env", func(t *testing.T) {
		t.Setenv(envTLSMinVersion, "1.0")
		// The environment variable is only a default, so a valid
		// attribute value is used and the invalid env value is ignored.
		require.NoError(t, validate(map[string]any{"tls_min_version": "1.2"}))
	})
}

func TestParseTLSMinVersion(t *testing.T) {
	v, err := parseTLSMinVersion("1.2")
	require.NoError(t, err)
	require.Equal(t, config.TLSVersion12, v)

	v, err = parseTLSMinVersion("1.3")
	require.NoError(t, err)
	require.Equal(t, config.TLSVersion13, v)

	for _, in := range append([]string{""}, invalidTLSMinVersions...) {
		_, err := parseTLSMinVersion(in)
		require.Error(t, err, in)
		require.Contains(t, err.Error(), `"1.2"`)
		require.Contains(t, err.Error(), `"1.3"`)
	}
}

func TestProvider_providerConfigure_tlsMinVersionInvalid(t *testing.T) {
	// TestResourceDataRaw skips schema validation, which covers values
	// that bypass it.
	d := schema.TestResourceDataRaw(t, Configure().Schema, map[string]any{"tls_min_version": "1.1"})
	_, diags := providerConfigure(d, nil)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Summary, "tls_min_version")
}

func TestProvider_providerConfigure_tlsMinVersion(t *testing.T) {
	tls12Server := newTLSServer(t, &tls.Config{MaxVersion: tls.VersionTLS12})
	tls13Server := newTLSServer(t, &tls.Config{MinVersion: tls.VersionTLS13})

	cases := []struct {
		name    string
		server  *httptest.Server
		attr    *string
		env     string
		wantErr bool
	}{
		{name: "tls 1.2 server, unset", server: tls12Server, wantErr: true},
		{name: "tls 1.2 server, 1.3", server: tls12Server, attr: ptr("1.3"), wantErr: true},
		{name: "tls 1.2 server, 1.2", server: tls12Server, attr: ptr("1.2")},
		{name: "tls 1.2 server, env 1.2", server: tls12Server, env: "1.2"},
		{name: "tls 1.2 server, env 1.3", server: tls12Server, env: "1.3", wantErr: true},
		{name: "tls 1.3 server, unset", server: tls13Server},
		{name: "tls 1.3 server, 1.2", server: tls13Server, attr: ptr("1.2")},
		{name: "tls 1.3 server, 1.3", server: tls13Server, attr: ptr("1.3")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(envTLSMinVersion, tc.env)

			raw := map[string]any{
				"remote_url":                tc.server.URL,
				"username":                  "admin",
				"password":                  "password",
				"tls_certificate_authority": writeServerCA(t, tc.server),
			}
			if tc.attr != nil {
				raw["tls_min_version"] = *tc.attr
			}

			// Validate as Terraform would before configuring the provider.
			diags := Provider().Validate(terraform.NewResourceConfigRaw(raw))
			require.False(t, diags.HasError(), "%v", diags)

			d := schema.TestResourceDataRaw(t, Configure().Schema, raw)
			meta, diags := providerConfigure(d, nil)
			require.False(t, diags.HasError(), "%v", diags)

			bp, ok := meta.(*client.BindPlane)
			require.True(t, ok)

			_, err := bp.Source("test")
			if tc.wantErr {
				require.Error(t, err)
				t.Logf("expected handshake error: %v", err)
				require.Contains(t, err.Error(), "protocol version not supported")
				return
			}
			require.NoError(t, err)
		})
	}
}

// newTLSServer starts an httptest TLS server with the given TLS config. It
// responds 404 to every request, which the client treats as a missing
// resource without an error.
func newTLSServer(t *testing.T, cfg *tls.Config) *httptest.Server {
	t.Helper()
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}))
	srv.TLS = cfg
	srv.StartTLS()
	t.Cleanup(srv.Close)
	return srv
}

// writeServerCA writes the server's certificate as PEM to a temp file and
// returns its path.
func writeServerCA(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ca.crt")
	data := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	require.NoError(t, os.WriteFile(path, data, 0o600))
	return path
}

// errorFromDiags joins the summary and detail of each diagnostic.
func errorFromDiags(diags diag.Diagnostics) error {
	errs := make([]error, 0, len(diags))
	for _, d := range diags {
		errs = append(errs, fmt.Errorf("%s: %s", d.Summary, d.Detail))
	}
	return errors.Join(errs...)
}

func ptr[T any](v T) *T { return &v }

func TestProvider_providerConfigure_tlsMinVersionHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	// tls_min_version only applies to https, plain http works with any value.
	for _, v := range []string{"1.2", "1.3"} {
		t.Run(v, func(t *testing.T) {
			raw := map[string]any{"remote_url": srv.URL, "tls_min_version": v}
			d := schema.TestResourceDataRaw(t, Configure().Schema, raw)
			meta, diags := providerConfigure(d, nil)
			require.False(t, diags.HasError(), "%v", diags)

			_, err := meta.(*client.BindPlane).Source("test")
			require.NoError(t, err)
		})
	}
}
