package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
)

func TestRun(t *testing.T) {
	i := build.Run{"echo %s %s", []string{"foo bar", "baz"}}

	assert.Equal(t, []string{`echo "foo bar" "baz"`}, i.Compile())
}

func TestRunWithInnerAndOuterArguments(t *testing.T) {
	i := build.Run{"useradd -d %s -u %s", []string{"/foo", "666", "bar"}}

	assert.Equal(t, []string{`useradd -d "/foo" -u "666" "bar"`}, i.Compile())
}

func TestRunAll(t *testing.T) {
	i := build.RunAll{[]build.Run{
		{"echo %s", []string{"foo"}},
		{"cat %s", []string{"/bar"}},
		{"baz", []string{}},
	}}

	assert.Equal(t, []string{`echo "foo" && cat "/bar" && baz`}, i.Compile())
}

func TestCopy(t *testing.T) {
	i := build.Copy{[]string{"source1", "source2"}, "dest"}

	assert.Equal(t, []string{`"source1"`, `"source2"`, `"dest/"`}, i.Compile())
}

func TestCopyAs(t *testing.T) {
	t.Run("wrapping Copy", func(t *testing.T) {
		i := build.CopyAs{
			123,
			124,
			build.Copy{[]string{"source1", "source2"}, "dest"},
		}

		assert.Equal(t, []string{"123:124", `"source1"`, `"source2"`, `"dest/"`}, i.Compile())
	})

	t.Run("wrapping CopyFrom", func(t *testing.T) {
		i := build.CopyAs{
			123,
			124,
			build.CopyFrom{"foo", build.Copy{[]string{"source1", "source2"}, "dest"}},
		}

		assert.Equal(t, []string{"123:124", "foo", `"source1"`, `"source2"`, `"dest/"`}, i.Compile())
	})
}

func TestCopyFrom(t *testing.T) {
	i := build.CopyFrom{"foo", build.Copy{[]string{"source1", "source2"}, "dest"}}

	assert.Equal(t, []string{"foo", `"source1"`, `"source2"`, `"dest/"`}, i.Compile())
}

func TestEntryPoint(t *testing.T) {
	i := build.EntryPoint{[]string{"/bin/foo", "bar", "baz"}}

	assert.Equal(t, []string{`"/bin/foo"`, `"bar"`, `"baz"`}, i.Compile())
}

func TestEnv(t *testing.T) {
	i := build.Env{map[string]string{
		"fooname": "foovalue",
		"barname": "barvalue",
		"quxname": "quxvalue",
	}}

	assert.Equal(t, []string{
		`barname="barvalue"`,
		`fooname="foovalue"`,
		`quxname="quxvalue"`,
	}, i.Compile())
}

func TestLabel(t *testing.T) {
	i := build.Label{map[string]string{
		"fooname": "foovalue",
		"barname": "barvalue",
		"quxname": "quxvalue",
	}}

	assert.Equal(t, []string{
		`barname="barvalue"`,
		`fooname="foovalue"`,
		`quxname="quxvalue"`,
	}, i.Compile())
}

func TestUser(t *testing.T) {
	i := build.User{"foo"}

	assert.Equal(t, []string{`"foo"`}, i.Compile())
}

func TestWorkingDirectory(t *testing.T) {
	i := build.WorkingDirectory{"/foo/path"}

	assert.Equal(t, []string{`"/foo/path"`}, i.Compile())
}
