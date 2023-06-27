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
	"github.com/observiq/terraform-provider-bindplane/internal/client"
)

const (
	envServerURL = "BINDPLANE_CONFIG_REMOTE_URL"
	envUsername  = "BINDPLANE_CONFIG_USERNAME" // #nosec, credentials are not hardcoded
	envPassword  = "BINDPLANE_CONFIG_PASSWORD" // #nosec, credentials are not hardcoded
	envTLSCa     = "BINDPLANE_CONFIG_TLS_CA"
	envTLSCrt    = "BINDPLANE_CONFIG_TLS_CERT"
	envTLSKey    = "BINDPLANE_CONFIG_TLS_KEY"

	// Timeout (including retries) for resources
	maxTimeout = time.Minute * 5
)

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"remote_url": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envServerURL,
				}, nil),
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envUsername,
				}, nil),
			},
			"password": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envPassword,
				}, nil),
			},
			"tls_certificate_authority": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envTLSCa,
				}, nil),
			},
			"tls_certificate": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envTLSCrt,
				}, nil),
			},
			"tls_private_key": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envTLSKey,
				}, nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"bindplane_configuration": resourceConfiguration(),
			"bindplane_destination":   resourceDestination(),
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		return providerConfigure(ctx, d, provider)
	}

	return provider
}

// providerConfigure configures the BindPlane client, which can be accessed from data / resource
// functions with with 'bindplane := meta.(client.BindPlane)'
func providerConfigure(_ context.Context, d *schema.ResourceData, _ *schema.Provider) (any, diag.Diagnostics) {
	var (
		endpoint string
		username string
		password string

		// tls
		ca   string
		cert string
		key  string
	)

	if v, ok := d.Get("remote_url").(string); ok {
		endpoint = v
	}

	if v, ok := d.Get("username").(string); ok {
		username = v
	}

	if v, ok := d.Get("password").(string); ok {
		password = v
	}

	if v, ok := d.Get("tls_certificate_authority").(string); ok {
		ca = v
	}

	if v, ok := d.Get("tls_certificate").(string); ok {
		cert = v
	}

	if v, ok := d.Get("tls_private_key").(string); ok {
		key = v
	}

	i, err := client.New(
		client.WithEndpoint(endpoint),
		client.WithUsername(username),
		client.WithPassword(password),

		// client side tls
		client.WithTLSTrustedCA(ca),

		// mutual tls
		client.WithTLS(cert, key),
	)
	if err != nil {
		err = fmt.Errorf("failed to initialize bindplane client: %w", err)
		return nil, diag.FromErr(err)
	}

	return i, nil
}
