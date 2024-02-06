package docindex

import (
	"encoding/json"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Meta contains the definition for a custom `x-docIndex` property use for
// explicitly ordering properties in documentation generated from JSON Schema.
var Meta = jsonschema.MustCompileString("docIndex.json", `{
	"properties" : {
		"x-docIndex": {
			"type": "integer"
		}
	}
}`)

// Compiler can be used to register the `x-docIndex` custom property with a
// jsonschema.Compiler.
type Compiler struct{}

// Compile detects use of `x-docIndex`
func (Compiler) Compile(ctx jsonschema.CompilerContext, m map[string]any) (jsonschema.ExtSchema, error) {
	if val, ok := m["x-docIndex"]; ok {
		idx, err := val.(json.Number).Int64()
		return Schema(idx), err
	}

	return nil, nil
}

// Schema represents the value of a `x-docIndex` property.
type Schema int64

// Validate is a noop as this property is used only in documentation
// generation.
func (Schema) Validate(ctx jsonschema.ValidationContext, v any) error {
	return nil
}
