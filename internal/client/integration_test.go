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

//go:build integration
// +build integration

package client

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"testing"
	"time"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/terraform-provider-bindplane/internal/configuration"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"

	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	bindplaneImage   = "ghcr.io/observiq/bindplane:1.17.1"
	bindplaneExtPort = 3100

	username = "int-test-user"
	password = "int-test-password"
)

func bindplaneContainer(t *testing.T, env map[string]string) testcontainers.Container {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	mount := testcontainers.ContainerMount{
		Source: testcontainers.GenericBindMountSource{
			HostPath: path.Join(dir, "tls"),
		},
		Target:   "/tmp",
		ReadOnly: false,
	}

	mounts := []testcontainers.ContainerMount{
		mount,
	}

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:  bindplaneImage,
		Env:    env,
		Mounts: mounts,
		// TODO(jsirianni): dynamic port?
		ExposedPorts: []string{fmt.Sprintf("%d:%d", bindplaneExtPort, 3001)},
		WaitingFor:   wait.ForListeningPort("3001"),
	}

	require.NoError(t, req.Validate())

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	time.Sleep(time.Second * 3)

	return container
}

func TestIntegration_http_raw_config(t *testing.T) {
	env := map[string]string{
		"BINDPLANE_USERNAME":       username,
		"BINDPLANE_PASSWORD":       password,
		"BINDPLANE_SESSION_SECRET": "524abde2-d9f8-485c-b426-bac229686d13",
		"BINDPLANE_SECRET_KEY":     "ED9B4232-C127-4580-9B86-62CEC420E7BB",
		"BINDPLANE_LOGGING_OUTPUT": "stdout",
	}

	container := bindplaneContainer(t, env)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()

	time.Sleep(time.Second * 20)

	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	endpoint := url.URL{
		Host:   fmt.Sprintf("%s:%d", hostname, bindplaneExtPort),
		Scheme: "http",
	}

	i, err := New(
		WithEndpoint(endpoint.String()),
		WithUsername(username),
		WithPassword(password),
	)
	require.NoError(t, err)

	_, err = i.client.Version(context.Background())
	require.NoError(t, err)

	_, err = i.client.Agents(context.Background(), client.QueryOptions{})
	require.NoError(t, err)

	config, err := configuration.NewV1(
		configuration.WithName("test"),
	)
	require.NoError(t, err)
	r := resource.AnyResourceFromRawConfigurationV1(config)
	require.NoError(t, i.Apply(&r, true))

	outputConfig, err := i.Configuration("test")
	require.NoError(t, err)
	require.NotNil(t, outputConfig)
	require.Equal(t, "test", outputConfig.Metadata.Name)

	err = i.DeleteConfiguration("test")
	require.NoError(t, err)

	outputConfig, err = i.Configuration("test")
	require.Nil(t, err, "reading a configuration that does not exist should return a nil error")
	require.Nil(t, outputConfig, "reading a configuration that does not exist should return a nil configuration")
}

func TestIntegration_http_config(t *testing.T) {
	env := map[string]string{
		"BINDPLANE_USERNAME":       username,
		"BINDPLANE_PASSWORD":       password,
		"BINDPLANE_SESSION_SECRET": "524abde2-d9f8-485c-b426-bac229686d13",
		"BINDPLANE_SECRET_KEY":     "ED9B4232-C127-4580-9B86-62CEC420E7BB",
		"BINDPLANE_LOGGING_OUTPUT": "stdout",
	}

	container := bindplaneContainer(t, env)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()

	time.Sleep(time.Second * 20)

	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	endpoint := url.URL{
		Host:   fmt.Sprintf("%s:%d", hostname, bindplaneExtPort),
		Scheme: "http",
	}

	i, err := New(
		WithEndpoint(endpoint.String()),
		WithUsername(username),
		WithPassword(password),
	)
	require.NoError(t, err)
	_, err = i.client.Version(context.Background())
	require.NoError(t, err)

	processorResource := model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       "Processor",
			Metadata: model.Metadata{
				Name: "my-processor",
			},
		},
		Spec: map[string]any{
			"type": "batch",
		},
	}
	require.NoError(t, i.ApplyWithRetry(
		context.TODO(),
		time.Duration(time.Minute*1),
		&processorResource, false), "did not expect an error when creating processor")
	_, err = i.Processor("my-processor")
	require.NoError(t, err)
	require.NoError(t, i.DeleteProcessor("my-processor"))

	sourceResource := model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       "Source",
			Metadata: model.Metadata{
				Name: "my-host",
			},
		},
		Spec: map[string]any{
			"type": "host",
		},
	}
	require.NoError(t, i.Apply(&sourceResource, false), "did not expect error when creating source")

	_, err = i.Source("my-host")
	require.NoError(t, err)

	destResource := model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       "Destination",
			Metadata: model.Metadata{
				Name: "logging",
			},
		},
		Spec: map[string]any{
			"type": "custom",
		},
	}
	require.NoError(t, i.Apply(&destResource, false), "did not expect error when creating destination")

	_, err = i.Destination("logging")
	require.NoError(t, err)

	// Missing resources should return nil because Terraform will take the
	// empty object and mark it as missing (to be created).
	_, err = i.Source("source-not-exist")
	require.NoError(t, err, "an error is not expected when looking up a source that does not exist")
	_, err = i.Destination("dest-not-exist")
	require.NoError(t, err, "an error is not expected when looking up a destination that does not exist")
	_, err = i.Processor("invalid-processor")
	require.NoError(t, err, "an error is not expected when looking up a processor that does not exist")

	// config params
	name := "test"
	labels := map[string]string{
		"purpose": "test",
	}
	matchLabels := map[string]string{
		"configuration": name,
	}
	sources := []configuration.ResourceConfig{
		{
			Name: "my-host",
		},
	}
	destinations := []configuration.ResourceConfig{
		{
			Name: "logging",
		},
	}

	config, err := configuration.NewV1(
		configuration.WithName(name),
		configuration.WithLabels(labels),
		configuration.WithMatchLabels(matchLabels),
		configuration.WithSourcesByName(sources),
		configuration.WithDestinationsByName(destinations),
	)
	require.NoError(t, err)
	r := resource.AnyResourceFromConfiguration(config)
	require.NoError(t, i.Apply(&r, true))

	config, err = i.Configuration(name)
	require.NoError(t, err)
	require.NotNil(t, config)
	require.Equal(t, name, config.Metadata.Name)
	require.Equal(t, labels, config.Metadata.Labels.AsMap())

	outputMatchLabels := make(map[string]string)
	for k, v := range config.Spec.Selector.MatchLabels {
		outputMatchLabels[k] = v
	}
	require.Equal(t, matchLabels, outputMatchLabels)

	err = i.DeleteSource("my-host")
	require.Error(t, err, "expected an error when deleting a source that has a dependent resource")

	err = i.DeleteDestination("logging")
	require.Error(t, err, "expected an error when deleting a destination that has a dependent resource")

	err = i.DeleteConfiguration("test")
	require.NoError(t, err)

	err = i.DeleteSource("my-host")
	require.NoError(t, err)

	err = i.DeleteDestination("logging")
	require.NoError(t, err)
}

func TestIntegration_invalidProtocol(t *testing.T) {
	env := map[string]string{
		"BINDPLANE_USERNAME":       username,
		"BINDPLANE_PASSWORD":       password,
		"BINDPLANE_SESSION_SECRET": "524abde2-d9f8-485c-b426-bac229686d13",
		"BINDPLANE_SECRET_KEY":     "ED9B4232-C127-4580-9B86-62CEC420E7BB",
		"BINDPLANE_LOGGING_OUTPUT": "stdout",
	}

	container := bindplaneContainer(t, env)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	endpoint := url.URL{
		Host:   fmt.Sprintf("%s:%d", hostname, bindplaneExtPort),
		Scheme: "https",
	}

	i, err := New(
		WithEndpoint(endpoint.String()),
		WithUsername(username),
		WithPassword(password),
		WithTLSTrustedCA("tls/bindplane-ca.crt"),
	)
	require.NoError(t, err)

	_, err = i.client.Version(context.Background())
	require.Error(t, err, "http: server gave HTTP response to HTTPS client")
}

func TestIntegration_https(t *testing.T) {
	env := map[string]string{
		"BINDPLANE_USERNAME":       username,
		"BINDPLANE_PASSWORD":       password,
		"BINDPLANE_TLS_CERT":       "/tmp/bindplane.crt",
		"BINDPLANE_TLS_KEY":        "/tmp/bindplane.key",
		"BINDPLANE_SESSION_SECRET": "524abde2-d9f8-485c-b426-bac229686d13",
		"BINDPLANE_SECRET_KEY":     "ED9B4232-C127-4580-9B86-62CEC420E7BB",
		"BINDPLANE_LOGGING_OUTPUT": "stdout",
	}

	container := bindplaneContainer(t, env)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	endpoint := url.URL{
		Host:   fmt.Sprintf("%s:%d", hostname, bindplaneExtPort),
		Scheme: "https",
	}

	i, err := New(
		WithEndpoint(endpoint.String()),
		WithUsername(username),
		WithPassword(password),
		WithTLSTrustedCA("tls/bindplane-ca.crt"),
	)
	require.NoError(t, err)

	_, err = i.client.Version(context.Background())
	require.NoError(t, err)

	_, err = i.client.Agents(context.Background(), client.QueryOptions{})
	require.NoError(t, err)
}

// func TestIntegration_mtls_fail(t *testing.T) {
// 	env := map[string]string{
// 		"BINDPLANE_USERNAME":       username,
// 		"BINDPLANE_PASSWORD":       password,
// 		"BINDPLANE_TLS_CERT":       "/tmp/bindplane.crt",
// 		"BINDPLANE_TLS_KEY":        "/tmp/bindplane.key",
// 		"BINDPLANE_TLS_CA":         "/tmp/bindplane-ca.crt",
// 		"BINDPLANE_SESSION_SECRET": "524abde2-d9f8-485c-b426-bac229686d13",
// 		"BINDPLANE_SECRET_KEY":     "ED9B4232-C127-4580-9B86-62CEC420E7BB",
// 		"BINDPLANE_LOGGING_OUTPUT": "stdout",
// 	}

// 	container := bindplaneContainer(t, env)
// 	defer func() {
// 		require.NoError(t, container.Terminate(context.Background()))
// 		time.Sleep(time.Second * 1)
// 	}()
// 	hostname, err := container.Host(context.Background())
// 	require.NoError(t, err)

// 	endpoint := url.URL{
// 		Host:   fmt.Sprintf("%s:%d", hostname, bindplaneExtPort),
// 		Scheme: "https",
// 	}

// 	i, err := New(
// 		WithEndpoint(endpoint.String()),
// 		WithUsername(username),
// 		WithPassword(password),
// 		WithTLSTrustedCA("tls/bindplane-ca.crt"),
// 	)
// 	require.NoError(t, err)

// 	_, err = i.client.Version(context.Background())
// 	require.Error(t, err, "expect an error when client not in mtls mode")
// 	require.Contains(t, err.Error(), "remote error: tls: bad certificate")
// }

func TestIntegration_mtls(t *testing.T) {
	env := map[string]string{
		"BINDPLANE_USERNAME":       username,
		"BINDPLANE_PASSWORD":       password,
		"BINDPLANE_TLS_CERT":       "/tmp/bindplane.crt",
		"BINDPLANE_TLS_KEY":        "/tmp/bindplane.key",
		"BINDPLANE_TLS_CA":         "/tmp/bindplane-ca.crt",
		"BINDPLANE_SESSION_SECRET": "524abde2-d9f8-485c-b426-bac229686d13",
		"BINDPLANE_SECRET_KEY":     "ED9B4232-C127-4580-9B86-62CEC420E7BB",
		"BINDPLANE_LOGGING_OUTPUT": "stdout",
	}

	container := bindplaneContainer(t, env)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	endpoint := url.URL{
		Host:   fmt.Sprintf("%s:%d", hostname, bindplaneExtPort),
		Scheme: "https",
	}

	i, err := New(
		WithEndpoint(endpoint.String()),
		WithUsername(username),
		WithPassword(password),
		WithTLSTrustedCA("tls/bindplane-ca.crt"),
		WithTLS("tls/bindplane-client.crt", "tls/bindplane-client.key"),
	)
	require.NoError(t, err)

	_, err = i.client.Version(context.Background())
	require.NoError(t, err)
}
