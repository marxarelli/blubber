package testtarget

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/util/llbtest"
)

// LLBAssertions aliases [llbtest.Assertions]
type LLBAssertions = llbtest.Assertions

// Assertions provides methods for making assertions on a [build.Target] and
// its marshaled [llb.Definition]. It extends [llbtest.Assertions] and
// [require.Assertions] to provide a singular assertions interface in
// [build.Target] related tests.
type Assertions struct {
	*LLBAssertions
	*require.Assertions

	t      *testing.T
	target *build.Target
}

// ExpandsEnv asserts that the expected string matches the latter string after
// the target substitutes previously set environment variables.
func (a *Assertions) ExpandsEnv(expected, expandable string) {
	a.t.Helper()
	a.Equal(expected, a.target.ExpandEnv(expandable))
}
