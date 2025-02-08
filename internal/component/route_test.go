// Copyright observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package component

import (
	"errors"
	"fmt"
	"testing"

	"github.com/observiq/bindplane-op-enterprise/model"
	"github.com/stretchr/testify/require"
)

func TestValidateRouteType(t *testing.T) {
	cases := []struct {
		routeType string
		expected  error
	}{
		{"logs", nil},
		{"metrics", nil},
		{"traces", nil},
		{"logs+metrics", nil},
		{"logs+traces", nil},
		{"metrics+traces", nil},
		{"logs+metrics+traces", nil},
		{"invalid", errors.New("invalid route type: invalid")},
	}

	for _, c := range cases {
		err := ValidateRouteType(c.routeType)
		if c.expected != nil {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestValidateRouteComponents(t *testing.T) {
	cases := []struct {
		components []model.ComponentPath
		expected   []error
	}{
		{[]model.ComponentPath{"destinations", "processors", "connectors"}, nil},
		{[]model.ComponentPath{"destinations", "processors", "invalid"}, []error{fmt.Errorf("invalid route component: invalid")}},
	}

	for _, c := range cases {
		errs := ValidateRouteComponents(c.components)
		if len(c.expected) == 0 {
			require.Equal(t, c.expected, errs)
		}
		require.Len(t, errs, len(c.expected))
	}
}
