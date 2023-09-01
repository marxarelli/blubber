package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

func TestApplyUser(t *testing.T) {
	instructions := []build.Instruction{
		build.Copy{[]string{"foo"}, "bar"},
		build.Copy{[]string{"baz"}, "qux"},
		build.CopyFrom{"foo", build.Copy{[]string{"a"}, "b"}},
	}

	assert.Equal(t,
		[]build.Instruction{
			build.CopyAs{"123", "223", build.Copy{[]string{"foo"}, "bar"}},
			build.CopyAs{"123", "223", build.Copy{[]string{"baz"}, "qux"}},
			build.CopyAs{"123", "223", build.CopyFrom{"foo", build.Copy{[]string{"a"}, "b"}}},
		},
		build.ApplyUser("123", "223", instructions),
	)
}

func TestChown(t *testing.T) {
	assert.Equal(
		t,
		build.Run{
			Command:   "chown %s:%s",
			Arguments: []string{"123", "124", "/foo"},
		},
		build.Chown("123", "124", "/foo"),
	)
}

func TestCreateDirectories(t *testing.T) {
	assert.Equal(
		t,
		build.Run{
			Command:   "mkdir -p",
			Arguments: []string{"/foo", "/bar"},
		},
		build.CreateDirectories([]string{"/foo", "/bar"}),
	)
}

func TestCreateDirectory(t *testing.T) {
	assert.Equal(
		t,
		build.Run{
			Command:   "mkdir -p",
			Arguments: []string{"/foo"},
		},
		build.CreateDirectory("/foo"),
	)
}

func TestCreateUser(t *testing.T) {
	assert.Equal(
		t,
		[]build.Run{
			{
				Command:   "(getent group %s || groupadd -o -g %s -r %s)",
				Arguments: []string{"124", "124", "foo"},
			},
			{
				Command:   "(getent passwd %s || useradd -l -o -m -d %s -r -g %s -u %s %s)",
				Arguments: []string{"123", "/home/foo", "124", "123", "foo"},
			},
		},
		build.CreateUser("foo", "123", "124"),
	)
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
