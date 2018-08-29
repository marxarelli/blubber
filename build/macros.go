package build

import (
	"fmt"
	"path"
	"sort"
)

// ApplyUser wraps any build.Copy instructions as build.CopyAs using the given
// UID/GID.
//
func ApplyUser(uid uint, gid uint, instructions []Instruction) []Instruction {
	applied := make([]Instruction, len(instructions))

	for i, instruction := range instructions {
		switch instruction.(type) {
		case Copy, CopyFrom:
			applied[i] = CopyAs{uid, gid, instruction}
		default:
			applied[i] = instruction
		}
	}

	return applied
}

// Chown returns a build.Run instruction for setting ownership on the given
// path.
//
func Chown(uid uint, gid uint, path string) Run {
	return Run{"chown %s:%s", []string{fmt.Sprint(uid), fmt.Sprint(gid), path}}
}

// CreateDirectories returns a build.Run instruction for creating all the
// given directories.
//
func CreateDirectories(paths []string) Run {
	return Run{"mkdir -p", paths}
}

// CreateDirectory returns a build.Run instruction for creating the given
// directory.
//
func CreateDirectory(path string) Run {
	return CreateDirectories([]string{path})
}

// CreateUser returns build.Run instructions for creating the given user
// account and group.
//
func CreateUser(name string, uid uint, gid uint) []Run {
	return []Run{
		{"groupadd -o -g %s -r", []string{fmt.Sprint(gid), name}},
		{"useradd -o -m -d %s -r -g %s -u %s", []string{homeDir(name), name, fmt.Sprint(uid), name}},
	}
}

// Home returns a build.Env instruction for setting the user's home directory.
//
func Home(name string) Env {
	return Env{map[string]string{"HOME": homeDir(name)}}
}

func homeDir(name string) string {
	if name == "root" {
		return "/root"
	}

	return "/home/" + name
}

// SortFilesByDir returns both the given files indexed by parent directory and
// a sorted slice of those parent directories. The latter is useful in
// ensuring deterministic iteration since the ordering of map keys is not
// guaranteed.
//
func SortFilesByDir(files []string) ([]string, map[string][]string) {
	bydir := make(map[string][]string)

	for _, file := range files {
		dir := path.Dir(file) + "/"
		file = path.Clean(file)

		if dirfiles, found := bydir[dir]; found {
			bydir[dir] = append(dirfiles, file)
		} else {
			bydir[dir] = []string{file}
		}
	}

	dirs := make([]string, len(bydir))
	i := 0

	for dir := range bydir {
		dirs[i] = dir
		i++
	}

	sort.Strings(dirs)

	return dirs, bydir
}

// SyncFiles returns build instructions to copy over the given files after
// creating their parent directories. Parent directories are created in a
// sorted order.
//
func SyncFiles(files []string, dest string) []Instruction {
	if len(files) < 1 {
		return []Instruction{}
	}

	dirs, bydir := SortFilesByDir(files)
	mkdirs := []string{}
	copies := make([]Instruction, len(dirs))

	// make project subdirectories for requirements files if necessary, and
	// copy in requirements files
	for i, dir := range dirs {
		fulldir := dest + "/" + dir
		fulldir = path.Clean(fulldir) + "/"

		if dir != "./" {
			mkdirs = append(mkdirs, fulldir)
		}

		copies[i] = Copy{bydir[dir], fulldir}
	}

	ins := []Instruction{}

	if len(mkdirs) > 0 {
		ins = append(ins, CreateDirectories(mkdirs))
	}

	return append(ins, copies...)
}
