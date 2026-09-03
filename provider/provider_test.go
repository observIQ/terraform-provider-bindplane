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

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/terraform-provider-bindplane/client"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	err := Provider().InternalValidate()
	require.NoError(t, err)
}

var _ *schema.Provider = Provider()

func TestProvider_providerConfigure(t *testing.T) {
	output, diag := providerConfigure(&schema.ResourceData{}, nil)
	require.Nil(t, diag)
	require.NotNil(t, output)

	// This test is critical because the provider relies heavily on type
	// assertion when interacting with the bindplane client
	i, ok := output.(*client.BindPlane)
	require.True(t, ok, "expected providerConfigure func to return type *bindplane.BindPlane")
	require.IsType(t, &client.BindPlane{}, i)
}

func TestProvider_providerConfigure_scopedKey(t *testing.T) {
	newData := func(raw map[string]any) *schema.ResourceData {
		return schema.TestResourceDataRaw(t, Configure().Schema, raw)
	}

	_, diags := providerConfigure(newData(map[string]any{"api_key": "bps_abc.def"}), nil)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Summary, "account_id")

	_, diags = providerConfigure(newData(map[string]any{"api_key": "bps_abc.def", "account_id": "01ABC"}), nil)
	require.False(t, diags.HasError())

	_, diags = providerConfigure(newData(map[string]any{"api_key": "legacy-key"}), nil)
	require.False(t, diags.HasError())

	_, diags = providerConfigure(newData(map[string]any{"api_key": "legacy-key", "account_id": "01ABC"}), nil)
	require.True(t, diags.HasError())

	_, diags = providerConfigure(newData(map[string]any{"username": "u", "password": "p", "account_id": "01ABC"}), nil)
	require.True(t, diags.HasError())
}
