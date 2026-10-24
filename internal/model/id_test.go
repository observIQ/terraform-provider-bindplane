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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewResourceID(t *testing.T) {
	before := time.Now().UnixMilli()
	id := NewResourceID()
	after := time.Now().UnixMilli()

	// IDs are 26 characters of Crockford's base32, the first of which
	// is at most 7 because the ID is 128 bits
	require.Regexp(t, `^[0-7][0-9A-HJKMNP-TV-Z]{25}$`, id)

	// The first 10 characters encode the millisecond timestamp
	var timestamp int64
	for _, c := range id[:10] {
		timestamp = timestamp<<5 | int64(strings.IndexRune(crockford, c))
	}
	require.GreaterOrEqual(t, timestamp, before)
	require.LessOrEqual(t, timestamp, after)
}

func TestNewResourceID_Unique(t *testing.T) {
	ids := map[string]bool{}
	for range 1000 {
		id := NewResourceID()
		require.False(t, ids[id], "duplicate id %s", id)
		ids[id] = true
	}
}
