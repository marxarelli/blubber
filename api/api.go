package api

import (
	// Embed our JSON Schema below
	_ "embed"
)

// ConfigSchema contains our embedded JSON Schema
//
//go:embed config.schema.json
var ConfigSchema string
