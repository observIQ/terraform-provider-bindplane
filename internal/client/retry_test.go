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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRetryableError(t *testing.T) {
	require.True(t, retryableError(errors.New("connect: connection refused")))
	require.True(t, retryableError(errors.New("Client.Timeout exceeded while awaiting headers")))
	require.False(t, retryableError(errors.New("")))
	require.False(t, retryableError(errors.New("resource is invalid")))
}
