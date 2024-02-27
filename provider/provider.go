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

// Package provider is the bindplane terraform provider.
package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ossClient "github.com/observiq/bindplane-op-enterprise/client"
	"github.com/observiq/bindplane-op-enterprise/config"
	"github.com/observiq/terraform-provider-bindplane/client"
	"go.uber.org/zap"
)

const (
	envAPIKey    = "BINDPLANE_TF_API_KEY" // #nosec G101 this is not a credential
	envRemoteURL = "BINDPLANE_TF_REMOTE_URL"
	envUsername  = "BINDPLANE_TF_USERNAME" // #nosec, credentials are not hardcoded
	envPassword  = "BINDPLANE_TF_PASSWORD" // #nosec, credentials are not hardcoded
	envTLSCa     = "BINDPLANE_TF_TLS_CA"
	envTLSCrt    = "BINDPLANE_TF_TLS_CERT"
	envTLSKey    = "BINDPLANE_TF_TLS_KEY"

	// Timeout (including retries) for resources
	maxTimeout = time.Minute * 5
)

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
	provider := Configure()

	provider.ConfigureContextFunc = func(_ context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		return providerConfigure(d, provider)
	}

	return provider
}

// Configure returns a configured provider with a schema.
func Configure() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"remote_url": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envRemoteURL,
				}, nil),
				Description: "The endpoint used to connect to the BindPlane OP instance.",
			},
			"api_key": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envAPIKey,
				}, nil),
				Description: "The API used to connect to the BindPlane OP instance.",
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envUsername,
				}, nil),
				Description: "The username used for authenticating to the BindPlane OP instance.",
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envPassword,
				}, nil),
				Description: "The password used for authenticating to the BindPlane OP instance.",
			},
			"tls_certificate_authority": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envTLSCa,
				}, nil),
				Description: "File path to the x509 PEM certificate authority file used for verifying the BindPlane OP instance's TLS certificate. Not required if your workstation already trusts the certificate authority through your operating system's trust store.",
			},
			"tls_certificate": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envTLSCrt,
				}, nil),
				Description: "File path to the x509 PEM client TLS certificate, required when the BindPlane OP instance is configured for mutual TLS.",
			},
			"tls_private_key": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envTLSKey,
				}, nil),
				Description: "File path to the x509 PEM client private key, required when the BindPlane OP instance is configured for mutual TLS.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"bindplane_configuration": resourceConfiguration(),
			"bindplane_destination":   resourceDestination(),
			"bindplane_source":        resourceSource(), // TODO(jsirianni): Determine if sources should be supported.
			"bindplane_processor":     resourceProcessor(),
			"bindplane_extension":     resourceExtension(),
		},
	}
}

// NewLogger returns a zap logger suitable for
// Terraform providers.
func NewLogger() (*zap.Logger, error) {
	loggerConf := zap.NewProductionConfig()
	loggerConf.OutputPaths = []string{"stdout"}
	logger, err := loggerConf.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to configure zap stdout logger: %w", err)
	}
	return logger, nil
}

// providerConfigure configures the BindPlane client, which can be accessed from data / resource
// functions with with 'bindplane := meta.(client.BindPlane)'
func providerConfigure(d *schema.ResourceData, _ *schema.Provider) (any, diag.Diagnostics) {
	config := &config.Config{}

	if v, ok := d.Get("api_key").(string); ok && v != "" {
		config.Auth.APIKey = v
	}

	if v, ok := d.Get("username").(string); ok && v != "" {
		config.Auth.Username = v
	}

	if v, ok := d.Get("password").(string); ok && v != "" {
		config.Auth.Password = v
	}

	if v, ok := d.Get("remote_url").(string); ok && v != "" {
		config.Network.RemoteURL = v
	}

	if v, ok := d.Get("tls_certificate_authority").(string); ok && v != "" {
		config.Network.TLS.CertificateAuthority = []string{v}
	}

	if crt, ok := d.Get("tls_certificate").(string); ok && crt != "" {
		if key, ok := d.Get("tls_private_key").(string); ok && key != "" {
			config.Network.TLS.Certificate = crt
			config.Network.TLS.PrivateKey = key
		}
	}

	logger, err := NewLogger()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	c, err := ossClient.NewBindPlane(config, logger)
	if err != nil {
		err = fmt.Errorf("failed to initialize bindplane client: %w", err)
		return nil, diag.FromErr(err)
	}

	return &client.BindPlane{
		Client: c,
	}, nil
}
