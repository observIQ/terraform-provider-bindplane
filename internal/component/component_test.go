package component

import (
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func TestNewResourceID(t *testing.T) {
	id := NewResourceID()
	require.True(t, strings.HasPrefix(id, "tf-"))
	_, err := ulid.Parse(strings.TrimPrefix(id, "tf-"))
	require.NoError(t, err)
}
