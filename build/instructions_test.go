package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/build"
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

	assert.Equal(t, []string{`"source1"`, `"source2"`, `"dest"`}, i.Compile())
}

func TestCopyFrom(t *testing.T) {
	i := build.CopyFrom{"foo", build.Copy{[]string{"source1", "source2"}, "dest"}}

	assert.Equal(t, []string{"foo", `"source1"`, `"source2"`, `"dest"`}, i.Compile())
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

func TestVolume(t *testing.T) {
	i := build.Volume{"/foo/dir"}

	assert.Equal(t, []string{`"/foo/dir"`}, i.Compile())
}
