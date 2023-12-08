// Package main exists to provide imports for tools used in development
package main

import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/google/addlicense"
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/mgechev/revive"
	_ "github.com/securego/gosec/v2/cmd/gosec"
	_ "github.com/uw-labs/lichen"
)
