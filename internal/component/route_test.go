package component

import (
	"errors"
	"fmt"
	"testing"

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
		components []string
		expected   []error
	}{
		{[]string{"destinations", "processors", "connectors"}, nil},
		{[]string{"destinations", "processors", "invalid"}, []error{fmt.Errorf("invalid route component: invalid")}},
	}

	for _, c := range cases {
		errs := ValidateRouteComponents(c.components)
		if len(c.expected) == 0 {
			require.Equal(t, c.expected, errs)
		}
		require.Len(t, errs, len(c.expected))
	}
}
