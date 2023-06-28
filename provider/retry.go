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

import "strings"

// Errors that should result in a retry are checked here
func retryableError(err error) bool {
	switch e := err.Error(); {
	case strings.Contains(e, "connect: connection refused"):
		return true
	case strings.Contains(e, "Client.Timeout exceeded while awaiting headers"):
		return true
	}
	return false
}
