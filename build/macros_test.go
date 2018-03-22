package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/build"
)

func TestApplyUser(t *testing.T) {
	instructions := []build.Instruction{
		build.Copy{[]string{"foo"}, "bar"},
		build.Copy{[]string{"baz"}, "qux"},
		build.CopyFrom{"foo", build.Copy{[]string{"a"}, "b"}},
	}

	assert.Equal(t,
		[]build.Instruction{
			build.CopyAs{123, 223, build.Copy{[]string{"foo"}, "bar"}},
			build.CopyAs{123, 223, build.Copy{[]string{"baz"}, "qux"}},
			build.CopyAs{123, 223, build.CopyFrom{"foo", build.Copy{[]string{"a"}, "b"}}},
		},
		build.ApplyUser(123, 223, instructions),
	)
}

func TestChown(t *testing.T) {
	i := build.Chown(123, 124, "/foo")

	assert.Equal(t, []string{`chown "123":"124" "/foo"`}, i.Compile())
}

func TestCreateDirectory(t *testing.T) {
	i := build.CreateDirectory("/foo")

	assert.Equal(t, []string{`mkdir -p "/foo"`}, i.Compile())
}

func TestCreateUser(t *testing.T) {
	i := build.CreateUser("foo", 123, 124)

	if assert.Len(t, i, 2) {
		assert.Equal(t, []string{`groupadd -o -g "124" -r "foo"`}, i[0].Compile())
		assert.Equal(t, []string{`useradd -o -m -d "/home/foo" -r -g "foo" -u "123" "foo"`}, i[1].Compile())
	}
}

func TestHome(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		assert.Equal(t,
			build.Env{map[string]string{"HOME": "/root"}},
			build.Home("root"),
		)
	})

	t.Run("non-root", func(t *testing.T) {
		assert.Equal(t,
			build.Env{map[string]string{"HOME": "/home/foo"}},
			build.Home("foo"),
		)
	})
}
