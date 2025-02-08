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
