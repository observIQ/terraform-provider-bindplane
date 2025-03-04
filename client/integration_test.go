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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/observiq/bindplane-op-enterprise/client"
	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/observiq/terraform-provider-bindplane/internal/configuration"
	"github.com/observiq/terraform-provider-bindplane/internal/resource"

	hashiversion "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	bindplaneExtPort = 3100

	username = "int-test-user"
	password = "int-test-password"

	networkName   = "bindplane-test-network"
	postgresName  = "bindplane-postgres"
	bindplaneName = "bindplane-server"
)

func createTestNetwork(t *testing.T, ctx context.Context) func(ctx context.Context) error {
	nw, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name: networkName,
		},
	})
	require.NoError(t, err)
	return nw.Remove
}

func bindplaneContainer(t *testing.T, ctx context.Context, env map[string]string, postgresHost string) (testcontainers.Container, *hashiversion.Version) {
	// Get the bindplane version to determine the image and tag
	version := os.Getenv("BINDPLANE_VERSION")
	if version == "" {
		t.Fatal("BINDPLANE_VERSION must be set: e.g. BINDPLANE_VERSION=v1.32.0")
	}

	// Trim the v prefix if not latest
	if version != "latest" {
		version = version[1:]
	}

	image := fmt.Sprintf("observiq/bindplane-ee:%s", version)

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	env["BINDPLANE_POSTGRES_HOST"] = postgresHost
	env["BINDPLANE_POSTGRES_PORT"] = "5432"
	env["BINDPLANE_POSTGRES_DATABASE"] = "bindplane"
	env["BINDPLANE_POSTGRES_USERNAME"] = "bindplane"
	env["BINDPLANE_POSTGRES_PASSWORD"] = "password"

	mount := testcontainers.ContainerMount{
		Source: testcontainers.GenericBindMountSource{
			HostPath: path.Join(dir, "tls"),
		},
		Target:   "/tmp",
		ReadOnly: false,
	}

	req := testcontainers.ContainerRequest{
		Image:  image,
		Env:    env,
		Name:   bindplaneName,
		Mounts: []testcontainers.ContainerMount{mount},
		// TODO(jsirianni): dynamic port?
		ExposedPorts: []string{fmt.Sprintf("%d:%d", bindplaneExtPort, 3001)},
		WaitingFor:   wait.ForListeningPort("3001"),
	}

	// Add network configuration if a network name is provided
	if networkName != "" {
		req.Networks = []string{networkName}
		req.NetworkAliases = map[string][]string{
			networkName: {bindplaneName},
		}
	}

	require.NoError(t, req.Validate())

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	time.Sleep(time.Second * 3)

	// Nil version implies latest
	var ver *hashiversion.Version
	if version != "latest" {
		v, err := hashiversion.NewVersion(version)
		if err != nil {
			container.Terminate(ctx)
			require.NoError(t, err, "failed to parse version")
		}
		ver = v
	}

	return container, ver
}

func bindplaneInit(endpoint url.URL, username, password string, version *hashiversion.Version) error {
	client := &http.Client{}

	switch endpoint.Scheme {
	case "http":
	case "https":
		clientCert, err := tls.LoadX509KeyPair("tls/bindplane-client.crt", "tls/bindplane-client.key")
		if err != nil {
			return fmt.Errorf("failed to load client cert: %w", err)
		}

		caCert, err := ioutil.ReadFile("tls/bindplane-ca.crt")
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{clientCert},
				RootCAs:      caCertPool,
			},
		}
	default:
		return fmt.Errorf("unsupported scheme: %s", endpoint.Scheme)
	}

	// 1.58.0 and older do not use organizations
	v158, err := hashiversion.NewVersion("1.58.0")
	if err != nil {
		return fmt.Errorf("failed to parse version 1.58.0: %w", err)
	}

	var data *strings.Reader
	if version == nil || version.Compare(v158) == 1 {
		endpoint.Path = "/v1/organizations"
		data = strings.NewReader(`{"organizationName": "init", "accountName": "project", "eulaAccepted":true}`)
	} else {
		endpoint.Path = "/v1/accounts"
		data = strings.NewReader(`{"displayName": "init"}`)
	}

	req, err := http.NewRequest("POST", endpoint.String(), data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var respBody map[string]interface{}
	return json.Unmarshal(body, &respBody)
}

func postgresContainer(t *testing.T, ctx context.Context, env map[string]string) (testcontainers.Container, string) {
	env["POSTGRES_PASSWORD"] = "password"
	env["POSTGRES_USER"] = "bindplane"
	env["POSTGRES_DB"] = "bindplane"

	postgresReq := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name:         postgresName,
			Image:        "postgres:16",
			Env:          env,
			ExposedPorts: []string{"5432:5432"},
			WaitingFor:   wait.ForListeningPort("5432"),
			Networks:     []string{networkName},
			NetworkAliases: map[string][]string{
				networkName: {postgresName},
			},
		},
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, postgresReq)
	require.NoError(t, err)

	// When using a network, we can just return the container name as the host
	return postgresContainer, postgresName
}

func TestIntegration_http_config(t *testing.T) {
	ctx := context.Background()
	networkCleanup := createTestNetwork(t, ctx)
	t.Cleanup(func() {
		require.NoError(t, networkCleanup(ctx))
	})

	license := os.Getenv("BINDPLANE_LICENSE")
	if license == "" {
		t.Fatal("BINDPLANE_LICENSE must be set in the environment")
	}

	env := map[string]string{
		"BINDPLANE_USERNAME":                      username,
		"BINDPLANE_PASSWORD":                      password,
		"BINDPLANE_SESSION_SECRET":                "524abde2-d9f8-485c-b426-bac229686d13",
		"BINDPLANE_LOGGING_OUTPUT":                "stdout",
		"BINDPLANE_ACCEPT_EULA":                   "true",
		"BINDPLANE_LICENSE":                       license,
		"BINDPLANE_TRANSFORM_AGENT_ENABLE_REMOTE": "true",
		"BINDPLANE_TRANSFORM_AGENT_REMOTE_AGENTS": "transform:4568",
		"BINDPLANE_POSTGRES_DATABASE":             "bindplane",
		"BINDPLANE_POSTGRES_USERNAME":             "bindplane",
		"BINDPLANE_POSTGRES_PASSWORD":             "password",
	}

	postgresContainer, postgresHost := postgresContainer(t, ctx, env)
	defer func() {
		require.NoError(t, postgresContainer.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()

	time.Sleep(time.Second * 1)
	err := postgresContainer.Start(ctx)
	require.NoError(t, err)

	// Create the bindplane container on the same network
	container, version := bindplaneContainer(t, ctx, env, postgresHost)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()

	time.Sleep(time.Second * 3)

	err = container.Start(ctx)
	time.Sleep(time.Minute * 1)
	require.NoError(t, err)

	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	endpoint := url.URL{
		Host:   net.JoinHostPort(hostname, fmt.Sprintf("%d", bindplaneExtPort)),
		Scheme: "http",
	}

	require.NoError(t, bindplaneInit(endpoint, username, password, version), "failed to initialize bindplane")

	i, err := newTestConfig(
		endpoint.String(),
		username,
		password,
		"", "", "",
	)
	require.NoError(t, err)
	_, err = i.Client.Version(context.Background())
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
		context.Background(),
		time.Duration(time.Minute*1),
		&processorResource, false), "did not expect an error when creating processor")
	_, err = i.GenericResource(model.KindProcessor, "my-processor")
	require.NoError(t, err)
	require.NoError(t, i.Delete(model.KindProcessor, "my-processor"))

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

	_, err = i.GenericResource(model.KindSource, "my-host")
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

	_, err = i.GenericResource(model.KindDestination, "logging")
	require.NoError(t, err)

	// Missing resources should return nil because Terraform will take the
	// empty object and mark it as missing (to be created).
	_, err = i.GenericResource(model.KindSource, "source-not-exist")
	require.NoError(t, err, "an error is not expected when looking up a source that does not exist")
	_, err = i.GenericResource(model.KindDestination, "dest-not-exist")
	require.NoError(t, err, "an error is not expected when looking up a destination that does not exist")
	_, err = i.GenericResource(model.KindProcessor, "invalid-processor")
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
	r := resource.AnyResourceFromConfigurationV1(config)
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

	err = i.Delete(model.KindSource, "my-host")
	require.Error(t, err, "expected an error when deleting a source that has a dependent resource")

	err = i.Delete(model.KindDestination, "logging")
	require.Error(t, err, "expected an error when deleting a destination that has a dependent resource")

	err = i.Delete(model.KindConfiguration, "test")
	require.NoError(t, err)

	err = i.Delete(model.KindSource, "my-host")
	require.NoError(t, err)

	err = i.Delete(model.KindDestination, "logging")
	require.NoError(t, err)

	err = i.Delete(model.KindAgent, "agent")
	require.Error(t, err, "Generic delete does not support agent")

	_, err = i.GenericResource(model.KindAgent, "agent")
	require.Error(t, err, "Generic get does not support agent")

	extensionsResource := model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       "Extension",
			Metadata: model.Metadata{
				Name: "my-extension",
			},
		},
		Spec: map[string]any{
			"type": "custom",
		},
	}

	require.NoError(t, i.Apply(&extensionsResource, false), "did not expect error when creating extension")
	require.NoError(t, i.Delete(model.KindExtension, "my-extension"), "did not expect error when deleting extension")
}

func TestIntegration_invalidProtocol(t *testing.T) {
	ctx := context.Background()
	networkCleanup := createTestNetwork(t, ctx)
	t.Cleanup(func() {
		require.NoError(t, networkCleanup(ctx))
	})

	license := os.Getenv("BINDPLANE_LICENSE")
	if license == "" {
		t.Fatal("BINDPLANE_LICENSE must be set in the environment")
	}

	env := map[string]string{
		"BINDPLANE_USERNAME":                      username,
		"BINDPLANE_PASSWORD":                      password,
		"BINDPLANE_SESSION_SECRET":                "524abde2-d9f8-485c-b426-bac229686d13",
		"BINDPLANE_LOGGING_OUTPUT":                "stdout",
		"BINDPLANE_ACCEPT_EULA":                   "true",
		"BINDPLANE_LICENSE":                       license,
		"BINDPLANE_TRANSFORM_AGENT_ENABLE_REMOTE": "true",
		"BINDPLANE_TRANSFORM_AGENT_REMOTE_AGENTS": "transform:4568",
	}

	// Create the postgres container on the network
	postgresContainer, postgresHost := postgresContainer(t, ctx, env)
	defer func() {
		require.NoError(t, postgresContainer.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()
	time.Sleep(time.Second * 1)
	err := postgresContainer.Start(ctx)
	require.NoError(t, err)

	container, version := bindplaneContainer(t, ctx, env, postgresHost)
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

	// Fix up the Scheme because this test purposefully uses the wrong scheme
	u := endpoint
	u.Scheme = "http"
	require.NoError(t, bindplaneInit(u, username, password, version), "failed to initialize bindplane")

	i, err := newTestConfig(
		endpoint.String(),
		username,
		password,
		"tls/bindplane-ca.crt", "", "",
	)
	require.NoError(t, err)

	_, err = i.Client.Version(context.Background())
	require.Error(t, err, "http: server gave HTTP response to HTTPS client")
}

func TestIntegration_https(t *testing.T) {
	ctx := context.Background()
	networkCleanup := createTestNetwork(t, ctx)
	t.Cleanup(func() {
		require.NoError(t, networkCleanup(ctx))
	})

	license := os.Getenv("BINDPLANE_LICENSE")
	if license == "" {
		t.Fatal("BINDPLANE_LICENSE must be set in the environment")
	}

	env := map[string]string{
		"BINDPLANE_USERNAME":                      username,
		"BINDPLANE_PASSWORD":                      password,
		"BINDPLANE_TLS_CERT":                      "/tmp/bindplane.crt",
		"BINDPLANE_TLS_KEY":                       "/tmp/bindplane.key",
		"BINDPLANE_SESSION_SECRET":                "524abde2-d9f8-485c-b426-bac229686d13",
		"BINDPLANE_LOGGING_OUTPUT":                "stdout",
		"BINDPLANE_ACCEPT_EULA":                   "true",
		"BINDPLANE_LICENSE":                       license,
		"BINDPLANE_TRANSFORM_AGENT_ENABLE_REMOTE": "true",
		"BINDPLANE_TRANSFORM_AGENT_REMOTE_AGENTS": "transform:4568",
		"BINDPLANE_POSTGRES_HOST":                 "postgres",
		"BINDPLANE_POSTGRES_PORT":                 "5432",
		"BINDPLANE_POSTGRES_DATABASE":             "bindplane",
		"BINDPLANE_POSTGRES_USERNAME":             "bindplane",
		"BINDPLANE_POSTGRES_PASSWORD":             "password",
	}

	// Create the postgres container on the network
	postgresContainer, postgresHost := postgresContainer(t, ctx, env)
	defer func() {
		require.NoError(t, postgresContainer.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()
	err := postgresContainer.Start(ctx)
	require.NoError(t, err)

	container, version := bindplaneContainer(t, ctx, env, postgresHost)
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

	require.NoError(t, bindplaneInit(endpoint, username, password, version), "failed to initialize bindplane")

	i, err := newTestConfig(
		endpoint.String(),
		username,
		password,
		"tls/bindplane-ca.crt", "", "",
	)
	require.NoError(t, err)

	_, err = i.Client.Version(context.Background())
	require.NoError(t, err)

	_, err = i.Client.Agents(context.Background(), client.QueryOptions{})
	require.NoError(t, err)
}

func TestIntegration_mtls(t *testing.T) {
	ctx := context.Background()
	networkCleanup := createTestNetwork(t, ctx)
	t.Cleanup(func() {
		require.NoError(t, networkCleanup(ctx))
	})

	license := os.Getenv("BINDPLANE_LICENSE")
	if license == "" {
		t.Fatal("BINDPLANE_LICENSE must be set in the environment")
	}

	env := map[string]string{
		"BINDPLANE_USERNAME":                      username,
		"BINDPLANE_PASSWORD":                      password,
		"BINDPLANE_TLS_CERT":                      "/tmp/bindplane.crt",
		"BINDPLANE_TLS_KEY":                       "/tmp/bindplane.key",
		"BINDPLANE_TLS_CA":                        "/tmp/bindplane-ca.crt",
		"BINDPLANE_SESSION_SECRET":                "524abde2-d9f8-485c-b426-bac229686d13",
		"BINDPLANE_LOGGING_OUTPUT":                "stdout",
		"BINDPLANE_ACCEPT_EULA":                   "true",
		"BINDPLANE_LICENSE":                       license,
		"BINDPLANE_TRANSFORM_AGENT_ENABLE_REMOTE": "true",
		"BINDPLANE_TRANSFORM_AGENT_REMOTE_AGENTS": "transform:4568",
		"BINDPLANE_POSTGRES_HOST":                 "postgres",
		"BINDPLANE_POSTGRES_PORT":                 "5432",
		"BINDPLANE_POSTGRES_DATABASE":             "bindplane",
		"BINDPLANE_POSTGRES_USERNAME":             "bindplane",
		"BINDPLANE_POSTGRES_PASSWORD":             "password",
	}
	// Create the postgres container on the network
	postgresContainer, postgresHost := postgresContainer(t, ctx, env)
	defer func() {
		require.NoError(t, postgresContainer.Terminate(context.Background()))
		time.Sleep(time.Second * 1)
	}()
	err := postgresContainer.Start(ctx)
	require.NoError(t, err)

	container, version := bindplaneContainer(t, ctx, env, postgresHost)
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

	require.NoError(t, bindplaneInit(endpoint, username, password, version), "failed to initialize bindplane")

	i, err := newTestConfig(
		endpoint.String(),
		username,
		password,
		"tls/bindplane-ca.crt",
		"tls/bindplane-client.crt",
		"tls/bindplane-client.key",
	)
	require.NoError(t, err)

	_, err = i.Client.Version(context.Background())
	require.NoError(t, err)
}
