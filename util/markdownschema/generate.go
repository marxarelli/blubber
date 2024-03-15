package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/atsushinee/go-markdown-generator/doc"
	"github.com/pborman/getopt"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gitlab.wikimedia.org/repos/releng/blubber/util/markdownschema/docindex"
)

const parameters = "schema.json"

var (
	showHelp = getopt.BoolLong("help", 'h', "show help/usage")
)

func main() {
	getopt.SetParameters(parameters)
	getopt.Parse()

	if *showHelp {
		getopt.Usage()
		os.Exit(1)
	}

	args := getopt.Args()

	if len(args) < 1 {
		getopt.Usage()
		os.Exit(1)
	}

	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true
	compiler.RegisterExtension("x-docIndex", docindex.Meta, docindex.Compiler{})

	schema := compiler.MustCompile(args[0])

	md := doc.NewMarkDown()

	err := writeSchema(md, "", schema, 1, false)
	if err != nil {
		log.Panic(err)
	}

	fmt.Print(md.String())
}

func writeSchema(md *doc.MarkDownDoc, name string, schema *jsonschema.Schema, lvl int, required bool) error {
	if schema == nil {
		return nil
	}

	if lvl > 4 {
		lvl = 4
	}

	schema = dereference(schema)
	title := titleOf(schema)

	if title != "" {
		md.WriteTitle(title, lvl)
	}

	subSchemas := append(schema.OneOf, schema.AnyOf...)
	if len(subSchemas) > 0 {
		for _, subSchema := range subSchemas {
			md.Writeln()
			writeSchema(md, name, subSchema, lvl, false)
		}
		return nil
	}

	if name != "" {
		nameAndType := fmt.Sprintf(
			"%s _%s_",
			md.GetCode(name),
			strings.Join(typesOf(schema), "|"),
		)
		if required {
			nameAndType += " (required)"
		}

		if title == "" {
			nameParts := strings.Split(name, ".")
			nameLast := nameParts[len(nameParts)-1]
			md.WriteTitle(nameLast, lvl)
		}

		md.Write(nameAndType)
		md.Writeln()
	}

	if schema.Description != "" {
		md.Writeln()
		md.Write(schema.Description)
		md.Writeln()
	}

	requiredProps := requiredOf(schema)
	names, props := propertiesOf(schema)

	for _, propName := range names {
		md.Writeln()
		writeSchema(md, name+"."+propName, props[propName], lvl+1, requiredProps[propName])
	}

	if additionalSchema, ok := schema.AdditionalProperties.(*jsonschema.Schema); ok {
		md.Writeln()
		writeSchema(md, name+".*", additionalSchema, lvl+1, false)
	}

	if schema.Items2020 != nil {
		md.Writeln()
		writeSchema(md, name+"[]", schema.Items2020, lvl+1, false)
	}

	return nil
}

func typesOf(schema *jsonschema.Schema) []string {
	schema = dereference(schema)

	types := schema.Types

	for i, t := range types {
		if t == "array" {
			if schema.Items2020 != nil {
				itemsTypes := typesOf(schema.Items2020)
				if len(itemsTypes) > 0 {
					types[i] = fmt.Sprintf("array&lt;%s&gt;", strings.Join(itemsTypes, "|"))
				}
			}
		}
	}

	for _, s := range schema.AllOf {
		types = append(types, typesOf(s)...)
	}

	for _, s := range schema.AnyOf {
		types = append(types, typesOf(s)...)
	}

	for _, s := range schema.OneOf {
		types = append(types, typesOf(s)...)
	}

	m := map[string]bool{}
	for _, t := range types {
		m[t] = true
	}

	uniq := make([]string, len(m))
	i := 0
	for t := range m {
		uniq[i] = t
		i++
	}

	sort.Strings(uniq)

	return uniq
}

func requiredOf(schema *jsonschema.Schema) map[string]bool {
	schema = dereference(schema)

	required := map[string]bool{}

	for _, name := range schema.Required {
		required[name] = true
	}

	for _, s := range schema.AllOf {
		for name, req := range requiredOf(s) {
			required[name] = req
		}
	}

	return required
}

func propertiesOf(schema *jsonschema.Schema) ([]string, map[string]*jsonschema.Schema) {
	schema = dereference(schema)

	props := map[string]*jsonschema.Schema{}

	for name, prop := range schema.Properties {
		props[name] = prop
	}

	for _, s := range schema.AllOf {
		_, nested := propertiesOf(s)
		for name, prop := range nested {
			props[name] = prop
		}
	}

	names := make([]string, len(props))
	i := 0
	for name := range props {
		names[i] = name
		i++
	}

	sort.SliceStable(
		names,
		func(i, j int) bool {
			a := props[names[i]]
			b := props[names[j]]

			aIdx := docIndexOf(a)
			bIdx := docIndexOf(b)

			if aIdx == bIdx {
				return names[i] < names[j]
			}

			return aIdx < bIdx
		},
	)

	return names, props
}

func titleOf(schema *jsonschema.Schema) string {
	schema = dereference(schema)

	title := schema.Title

	for _, s := range schema.AllOf {
		if s.Title != "" {
			title = s.Title
		}
	}

	return title
}

func docIndexOf(schema *jsonschema.Schema) int64 {
	schema = dereference(schema)

	if ext, ok := schema.Extensions["x-docIndex"]; ok {
		if docIndex, ok := ext.(docindex.Schema); ok {
			return int64(docIndex)
		}
	}

	return 0
}

func dereference(schema *jsonschema.Schema) *jsonschema.Schema {
	if schema.Ref == nil {
		return schema
	}

	return dereference(schema.Ref)
}
