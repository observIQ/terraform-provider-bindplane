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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

// Minimum TLS versions accepted by Config.TLSMinVersion.
const (
	TLSVersion12 = "1.2"
	TLSVersion13 = "1.3"
)

// newTLSConfig returns the TLS configuration used to connect to Bindplane.
// It matches the Bindplane client: TLS 1.3 is the minimum version unless
// lowered with TLSMinVersion, and a configured certificate authority
// replaces the system's trusted roots.
func newTLSConfig(cfg Config) (*tls.Config, error) {
	minVersion, err := parseTLSVersion(cfg.TLSMinVersion)
	if err != nil {
		return nil, err
	}

	// #nosec G402 skip verify is opt in, and documented as only for testing
	tlsConfig := &tls.Config{
		MinVersion:         minVersion,
		InsecureSkipVerify: cfg.TLSSkipVerify,
	}

	if cfg.TLSCertificateAuthority != "" {
		ca, err := os.ReadFile(cfg.TLSCertificateAuthority) // #nosec G304 user defines the ca file path
		if err != nil {
			return nil, fmt.Errorf("failed to read certificate authority file: %w", err)
		}

		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(ca) {
			return nil, errors.New("failed to append certificate authority to root ca pool")
		}
		tlsConfig.RootCAs = pool
	}

	if cfg.TLSCertificate != "" && cfg.TLSPrivateKey != "" {
		keypair, err := tls.LoadX509KeyPair(cfg.TLSCertificate, cfg.TLSPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load tls certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{keypair}
	}

	return tlsConfig, nil
}

// parseTLSVersion converts a TLSMinVersion value to a crypto/tls version.
func parseTLSVersion(v string) (uint16, error) {
	switch v {
	case "", TLSVersion13:
		return tls.VersionTLS13, nil
	case TLSVersion12:
		return tls.VersionTLS12, nil
	default:
		return 0, fmt.Errorf("invalid TLS min version %q, must be one of %q", v, []string{TLSVersion12, TLSVersion13})
	}
}
