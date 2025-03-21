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

// Package component provides functions for defining bindplane
// component values.
package component

import (
	"fmt"

	"github.com/observiq/bindplane-op-enterprise/model"
)

// NewResourceID wraps model.NewResourceID and returns
// a new resource ID with the `tf` prefix to indicate
// that it was created by the Terraform provider.
func NewResourceID() string {
	return fmt.Sprintf("tf-%s", model.NewResourceID())
}
