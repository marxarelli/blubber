package build_test

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"

	"phabricator.wikimedia.org/source/blubber.git/build"
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
