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
