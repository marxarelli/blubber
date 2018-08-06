package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
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

func TestCreateDirectories(t *testing.T) {
	i := build.CreateDirectories([]string{"/foo", "/bar"})

	assert.Equal(t, []string{`mkdir -p "/foo" "/bar"`}, i.Compile())
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

func TestSortFilesByDir(t *testing.T) {
	files := []string{"foo", "./bar", "./d/d-foo", "./c/c/c-foo", "b/b-foo", "b/b-bar", "a/a-foo"}

	sortedDirs, filesByDir := build.SortFilesByDir(files)

	assert.Equal(t,
		[]string{
			"./",
			"a/",
			"b/",
			"c/c/",
			"d/",
		},
		sortedDirs,
	)

	assert.Equal(t,
		map[string][]string{
			"./":   []string{"foo", "bar"},
			"d/":   []string{"d/d-foo"},
			"c/c/": []string{"c/c/c-foo"},
			"b/":   []string{"b/b-foo", "b/b-bar"},
			"a/":   []string{"a/a-foo"},
		},
		filesByDir,
	)
}

func TestSyncFiles(t *testing.T) {
	files := []string{"foo", "./bar", "./d/d-foo", "./c/c/c-foo", "b/b-foo", "b/b-bar", "a/a-foo"}

	assert.Equal(t,
		[]build.Instruction{
			build.Run{"mkdir -p", []string{"a/", "b/", "c/c/", "d/"}},
			build.Copy{[]string{"foo", "bar"}, "./"},
			build.Copy{[]string{"a/a-foo"}, "a/"},
			build.Copy{[]string{"b/b-foo", "b/b-bar"}, "b/"},
			build.Copy{[]string{"c/c/c-foo"}, "c/c/"},
			build.Copy{[]string{"d/d-foo"}, "d/"},
		},
		build.SyncFiles(files, "."),
	)
}

func TestSyncFilesWithDestination(t *testing.T) {
	files := []string{"foo", "./bar", "./d/d-foo", "./c/c/c-foo", "b/b-foo", "b/b-bar", "a/a-foo"}

	assert.Equal(t,
		[]build.Instruction{
			build.Run{"mkdir -p", []string{"/dir/a/", "/dir/b/", "/dir/c/c/", "/dir/d/"}},
			build.Copy{[]string{"foo", "bar"}, "/dir/"},
			build.Copy{[]string{"a/a-foo"}, "/dir/a/"},
			build.Copy{[]string{"b/b-foo", "b/b-bar"}, "/dir/b/"},
			build.Copy{[]string{"c/c/c-foo"}, "/dir/c/c/"},
			build.Copy{[]string{"d/d-foo"}, "/dir/d/"},
		},
		build.SyncFiles(files, "/dir"),
	)
}
