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
	"fmt"
	"net/http"
	"strings"
)

// Config is the configuration used to connect to Bindplane.
type Config struct {
	// RemoteURL is the Bindplane endpoint, e.g. https://bindplane.example.com
	RemoteURL string

	// APIKey authenticates with an API key. It takes precedence
	// over Username and Password.
	APIKey string

	// AccountID is sent as the X-Bindplane-Account-ID header.
	AccountID string

	// Username and Password authenticate with basic auth when
	// both are set.
	Username string
	Password string

	// TLSCertificateAuthority is the path to a PEM certificate authority
	// used to verify the Bindplane server's certificate.
	TLSCertificateAuthority string

	// TLSCertificate and TLSPrivateKey are paths to a PEM client
	// certificate and private key, used for mutual TLS.
	TLSCertificate string
	TLSPrivateKey  string

	// TLSSkipVerify disables verification of the server's certificate.
	TLSSkipVerify bool
}

// New returns a BindPlane client configured with cfg.
func New(cfg Config) (*BindPlane, error) {
	tlsConfig, err := newTLSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("configure TLS client: %w", err)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig

	return &BindPlane{
		client: &http.Client{
			Timeout:   defaultTimeout,
			Transport: transport,
		},
		baseURL: strings.TrimSuffix(cfg.RemoteURL, "/") + "/v1",
		config:  cfg,
	}, nil
}
