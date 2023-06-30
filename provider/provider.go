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
	envServerURL = "BINDPLANE_TF_REMOTE_URL"
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
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the bindplane client profile in ~/.bindplane. All other configuration options will override values set by the profile.",
			},
			"remote_url": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					envServerURL,
				}, nil),
				Description: "The endpoint used to connect to the BindPlane OP instance.",
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
			"bindplane_configuration":     resourceConfiguration(),
			"bindplane_raw_configuration": resourceRawConfiguration(),
			"bindplane_destination":       resourceDestination(),
			"bindplane_source":            resourceSource(), // TODO(jsirianni): Determine if sources should be supported.
			"bindplane_processor":         resourceProcessor(),
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
	profile := ""
	if v, ok := d.Get("profile").(string); ok {
		profile = v
	}

	clientOpts := []client.Option{}

	if v, ok := d.Get("remote_url").(string); ok && v != "" {
		clientOpts = append(clientOpts, client.WithEndpoint(v))
	}

	if v, ok := d.Get("username").(string); ok && v != "" {
		clientOpts = append(clientOpts, client.WithUsername(v))
	}

	if v, ok := d.Get("password").(string); ok && v != "" {
		clientOpts = append(clientOpts, client.WithPassword(v))
	}

	if v, ok := d.Get("tls_certificate_authority").(string); ok && v != "" {
		clientOpts = append(clientOpts, client.WithTLSTrustedCA(v))
	}

	if crt, ok := d.Get("tls_certificate").(string); ok && crt != "" {
		if key, ok := d.Get("tls_private_key").(string); ok && key != "" {
			clientOpts = append(clientOpts, client.WithTLS(crt, key))
		}
	}

	i, err := client.New(profile, clientOpts...)
	if err != nil {
		err = fmt.Errorf("failed to initialize bindplane client: %w", err)
		return nil, diag.FromErr(err)
	}

	return i, nil
}
