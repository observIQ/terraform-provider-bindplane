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

	"github.com/observiq/bindplane-op-enterprise/client"
	"github.com/observiq/bindplane-op-enterprise/config"
	"go.uber.org/zap"
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

	// TLSMinVersion is the minimum TLS version, "1.2" or "1.3". The
	// Bindplane client default is used when empty.
	TLSMinVersion string
}

// New returns a BindPlane client configured with cfg.
func New(cfg Config) (*BindPlane, error) {
	c := &config.Config{}
	c.Network.RemoteURL = cfg.RemoteURL
	c.Auth.APIKey = cfg.APIKey
	c.Accounts.Account = cfg.AccountID
	c.Auth.Username = cfg.Username
	c.Auth.Password = cfg.Password

	// System auth type must be set when using user / pass. When using API
	// key, auth type does not need to be set client side.
	if cfg.Username != "" && cfg.Password != "" {
		c.Auth.Type = "system"
	}

	if cfg.TLSCertificateAuthority != "" {
		c.Network.TLS.CertificateAuthority = []string{cfg.TLSCertificateAuthority}
	}

	if cfg.TLSCertificate != "" && cfg.TLSPrivateKey != "" {
		c.Network.TLS.Certificate = cfg.TLSCertificate
		c.Network.TLS.PrivateKey = cfg.TLSPrivateKey
	}

	c.Network.TLS.InsecureSkipVerify = cfg.TLSSkipVerify

	if cfg.TLSMinVersion != "" {
		c.Network.TLS.MinTLSVersion = config.TLSVersion(cfg.TLSMinVersion)
	}

	loggerConf := zap.NewProductionConfig()
	loggerConf.OutputPaths = []string{"stdout"}
	logger, err := loggerConf.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to configure zap stdout logger: %w", err)
	}

	bp, err := client.NewBindplane(c, logger)
	if err != nil {
		return nil, err
	}

	return &BindPlane{Client: bp}, nil
}
