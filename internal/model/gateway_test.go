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

package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGatewayMode(t *testing.T) {
	var g Gateway
	require.NoError(t, json.Unmarshal([]byte(`{"kind": "Gateway", "spec": {"configuration": "c", "loadBalancerHostname": "lb"}}`), &g))
	require.Equal(t, GatewayModeManual, g.Mode())
	require.Equal(t, "c", g.GatewaySpec.Configuration)
	require.Equal(t, "lb", g.GatewaySpec.LoadBalancerHostname)

	require.NoError(t, json.Unmarshal([]byte(`{"kind": "Gateway", "spec": {"workload": {"operator": "o", "workload": "w"}}}`), &g))
	require.Equal(t, GatewayModeOperator, g.Mode())
}

func TestTrimVersion(t *testing.T) {
	require.Equal(t, "name", TrimVersion("name"))
	require.Equal(t, "name", TrimVersion("name:2"))
	require.Equal(t, "name", TrimVersion("name:latest"))
	require.Equal(t, "name", TrimVersion("name:"))
}
