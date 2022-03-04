//go:build tools
// +build tools

package main

// This file maintains references to go modules needed during development and
// testing to keep `go mod tidy` from removing their go.mod entries.
import (
	_ "golang.org/x/lint/golint"
)
