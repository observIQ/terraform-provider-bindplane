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
	"errors"
	"fmt"
	"testing"

	"github.com/observiq/bindplane-op-enterprise/client"
	"github.com/observiq/bindplane-op-enterprise/config"
	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestConfig(endpoint, user, pass, ca, crt, key string) (*BindPlane, error) {
	config := &config.Config{}
	var err error

	config.Network.RemoteURL = endpoint
	config.Auth.Username = user
	config.Auth.Password = pass

	// System auth type must be set when using user / pass. When using API
	// key, auth type does not need to be set client side.
	if config.Auth.Username != "" && config.Auth.Password != "" {
		config.Auth.Type = "system"
	}

	if ca != "" {
		config.Network.TLS.CertificateAuthority = []string{ca}
	}
	if crt != "" && key != "" {
		config.Network.TLS.Certificate = crt
		config.Network.TLS.PrivateKey = key
	}

	loggerConf := zap.NewProductionConfig()
	loggerConf.OutputPaths = []string{"stdout"}
	logger, err := loggerConf.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to configure zap stdout logger: %w", err)
	}

	i, err := client.NewBindPlane(config, logger)
	if err != nil {
		return nil, err
	}

	return &BindPlane{i}, nil
}

func TestNewTestConfig(t *testing.T) {
	c, err := newTestConfig(
		"http://go.dev",
		"otelu",
		"otelp",
		"",
		"",
		"",
	)
	require.NoError(t, err)
	require.NotNil(t, c)
}

// BindPlane is not configured, API calls should fail
func TestApply(t *testing.T) {
	i, err := newTestConfig("", "", "", "", "", "")
	require.NoError(t, err)
	require.NotNil(t, i)
	require.Error(t, i.Apply(&model.AnyResource{}, false))
}

// BindPlane is not configured, API calls should fail
func TestConfiguration(t *testing.T) {
	i, err := newTestConfig("", "", "", "", "", "")
	require.NoError(t, err)
	require.NotNil(t, i)

	_, err = i.Configuration("does-not-exist")
	require.Error(t, err)
}

// BindPlane is not configured, API calls should fail
func TestDeleteConfiguration(t *testing.T) {
	i, err := newTestConfig("", "", "", "", "", "")
	require.NoError(t, err)
	require.NotNil(t, i)

	err = i.DeleteConfiguration("does-not-exist")
	require.Error(t, err)
}

func TestIsNotFoundError(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		expect bool
	}{
		{
			"upper-true",
			errors.New("404 Not Found"),
			true,
		},
		{
			"lower-true",
			errors.New("404 not found"),
			true,
		},
		{
			"false",
			errors.New("not found"),
			false,
		},
		{
			"false",
			errors.New("404"),
			false,
		},
		{
			"false",
			errors.New("error"),
			false,
		},
	}

	for _, tc := range cases {
		out := isNotFoundError(tc.err)
		require.Equal(t, tc.expect, out)
	}
}
