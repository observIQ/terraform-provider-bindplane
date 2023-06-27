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

// Package client provides a client wrapper for bindplane-op/client.NewBindPlane.
// It provides wrapper functions suitable for Terraform use. For example, when
// BindPlane server returns a 404, this wrapper package will gracefully handle it
// instead of returning an error.
package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/config"
	"github.com/observiq/bindplane-op/model"
	"go.uber.org/zap"
)

// New takes configuration options and returns a BindPlane client.
func New(options ...Option) (*BindPlane, error) {
	config := &config.Config{}

	for _, option := range options {
		if option != nil {
			option(config)
		}
	}

	loggerConf := zap.NewProductionConfig()
	loggerConf.OutputPaths = []string{"stdout"}
	logger, err := loggerConf.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to configure zap stdout logger: %w", err)
	}

	i, err := client.NewBindPlane(config, logger)
	return &BindPlane{i}, err
}

// Option is a function that configures a BindPlane client configuration
type Option func(*config.Config)

// WithEndpoint sets a client's endpoint
func WithEndpoint(endpoint string) Option {
	return func(c *config.Config) {
		c.Network.RemoteURL = endpoint
	}
}

// WithUsername sets a client's username
func WithUsername(username string) Option {
	return func(c *config.Config) {
		c.Auth.Username = username
	}
}

// WithPassword sets a client's password
func WithPassword(password string) Option {
	return func(c *config.Config) {
		c.Auth.Password = password
	}
}

// WithTLSTrustedCA sets a client's trusted
// certificate authorities. A single terraform
// client can talk to many BindPlane servers with
// different certificate authorities.
func WithTLSTrustedCA(path string) Option {
	if path == "" {
		return nil
	}

	return func(c *config.Config) {
		c.Network.CertificateAuthority = []string{path}
	}
}

// WithTLS sets a client's TLS certificate and key file
func WithTLS(crt, key string) Option {
	if crt == "" || key == "" {
		return nil
	}

	return func(c *config.Config) {
		c.Network.Certificate = crt
		c.Network.PrivateKey = key
	}
}

// BindPlane is a shim layer between Terraform and the
// BindPlane client interface
type BindPlane struct {
	client client.BindPlane
}

// Apply creates or updates a single BindPlane resource and returns it's id.
func (i *BindPlane) Apply(r *model.AnyResource) (string, error) {
	status, err := i.client.Apply(context.TODO(), []*model.AnyResource{r})
	if err != nil {
		return "", fmt.Errorf("failed to apply BindPlane resources: %w", err)
	}

	// BindPlane should return a single status when applying a single resource,
	// which means we will index into status[0] when returning the id.
	if x := len(status); x != 1 {
		return "", fmt.Errorf("expected BindPlane to return one resource status, got %d", x)
	}

	resource := status[0]

	// Apply expects the resource to be unchanged, configured, or
	// created. All other statuses are unexpected and should result
	// in an error from this method.
	switch resource.Status {
	case model.StatusUnchanged, model.StatusConfigured, model.StatusCreated:
		break
	default:
		return "", fmt.Errorf(
			"unexpected status when applying resource: %s, status: %s",
			r.Name(),
			resource.Status)
	}

	id := status[0].Resource.ID()
	return id, nil
}

// Configuration takes a name and returns the matching configuration
func (i *BindPlane) Configuration(name string) (*model.Configuration, error) {
	c, err := i.client.Configuration(context.TODO(), name)
	if err != nil {
		// Do not return an error if the resource is not found. Terraform
		// will understand that the resource does not exist when it receives
		// a nil value, and will instead offer to create the resource.
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get configuration with name %s: %w", name, err)
	}
	return c, nil
}

// DeleteConfiguration will delete a BindPlane configuration
func (i *BindPlane) DeleteConfiguration(name string) error {
	err := i.client.DeleteConfiguration(context.TODO(), name)
	if err != nil {
		return fmt.Errorf("error while deleting configuration with name %s: %w", name, err)
	}
	return nil
}

// Destination takes a name and returns the matching destination
func (i *BindPlane) Destination(name string) (*model.Destination, error) {
	r, err := i.client.Destination(context.TODO(), name)
	if err != nil {
		// Do not return an error if the resource is not found. Terraform
		// will understand that the resource does not exist when it receives
		// a nil value, and will instead offer to create the resource.
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get destination with name %s: %v", name, err)
	}
	return r, nil
}

// DeleteDestination will delete a BindPlane destination
func (i *BindPlane) DeleteDestination(name string) error {
	err := i.client.DeleteDestination(context.TODO(), name)
	if err != nil {
		return fmt.Errorf("error while deleting destination with name %s: %w", name, err)
	}
	return nil
}

// TODO(jsirianni): BindPlane should probably have error types so we can check
// error.Is.
func isNotFoundError(err error) bool {
	e := strings.ToLower(err.Error())
	return strings.Contains(e, "404 not found")
}
