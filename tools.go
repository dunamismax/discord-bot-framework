//go:build tools

// Package tools manages tool dependencies for the Discord Bot Framework.
package tools

import (
	// Development and build tools
	_ "github.com/magefile/mage"
	
	// Linting and code quality tools (installed via go install)
	// _ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	// _ "golang.org/x/tools/cmd/goimports"
	// _ "golang.org/x/vuln/cmd/govulncheck"
	
	// Testing tools
	// _ "github.com/stretchr/testify"
)