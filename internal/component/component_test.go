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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewResourceID(t *testing.T) {
	// IDs are "tf-" followed by a ULID: 26 characters of Crockford's
	// base32, the first of which is at most 7.
	const pattern = `^tf-[0-7][0-9A-HJKMNP-TV-Z]{25}$`

	ids := map[string]bool{}
	for range 1000 {
		id := NewResourceID()
		require.Regexp(t, pattern, id)
		require.False(t, ids[id], "duplicate id %s", id)
		ids[id] = true
	}
}
