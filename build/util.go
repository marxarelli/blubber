package build

import (
	"fmt"
	"sort"
	"strconv"
)

func sortedKeys(keyValues map[string]string) []string {
	keys := make([]string, len(keyValues))

	i := 0
	for k := range keyValues {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	return keys
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
