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
	"context"
	"fmt"
	"time"

	tfresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/observiq/terraform-provider-bindplane/internal/client"
)

// resourceGenericConfigurationDelete can delete configurations and raw configurations.
func resourceGenericConfigurationDelete(d *schema.ResourceData, meta any) error {
	bindplane := meta.(*client.BindPlane)

	err := tfresource.RetryContext(context.TODO(), d.Timeout(schema.TimeoutDelete)-time.Minute, func() *tfresource.RetryError {
		name := d.Get("name").(string)
		err := bindplane.DeleteConfiguration(name)
		if err != nil {
			err := fmt.Errorf("failed to delete configuration '%s' by name: %v", name, err)
			if retryableError(err) {
				return tfresource.RetryableError(err)
			}
			return tfresource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("delete retries exhausted: %v", err)
	}

	return resourceConfigurationRead(d, meta)
}

func isValidPlatform(platform string) bool {
	// TODO(jsirianni): We should use a bindplane-op package to determine
	// valid platforms.
	const (
		platformWindows             = "windows"
		platformLinux               = "linux"
		platformMacOS               = "macos"
		platformK8sDaemonset        = "kubernetes-daemonset"
		platformK8sDeployment       = "kubernetes-deployment"
		platformOpenshiftDaemonset  = "openshift-daemonset"
		platformOpenshiftDeployment = "openshift-deployment"
	)
	switch platform {
	case platformWindows, platformLinux, platformMacOS,
		platformK8sDaemonset, platformK8sDeployment,
		platformOpenshiftDaemonset, platformOpenshiftDeployment:
		return true
	default:
		return false
	}
}
