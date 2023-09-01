package build

import (
	"fmt"
	"sort"
	"strconv"
)

func sortedEnvKeyValues(keyValues map[string]string) []string {
	defs := make([]string, 0, len(keyValues))
	names := make([]string, 0, len(keyValues))

	for name := range keyValues {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		defs = append(defs, name+"="+quote(keyValues[name]))
	}

	return defs
}

func quote(arg string) string {
	return strconv.Quote(arg)
}

func quoteAll(arguments []string) []string {
	quoted := make([]string, len(arguments))

	for i, arg := range arguments {
		quoted[i] = quote(arg)
	}

	return quoted
}

func sprintf(format string, arguments []string) string {
	args := make([]interface{}, len(arguments))

	for i, v := range arguments {
		args[i] = quote(v)
	}

	return fmt.Sprintf(format, args...)
}
