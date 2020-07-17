package config_test

import (
	"testing"

	"gerrit.wikimedia.org/r/blubber/config"
	"github.com/stretchr/testify/assert"
)

func TestDepGraphGood(t *testing.T) {
	graph := config.NewDepGraph()

	graph.AddDependency("top", "mid1")
	graph.AddDependency("top", "mid2")
	graph.AddDependency("mid1", "leaf1")
	graph.AddDependency("mid1", "leaf2")
	graph.AddDependency("mid2", "leaf2")
	graph.AddDependency("mid2", "leaf3")

	res, err := graph.GetDeps("top")

	assert.Nil(t, err, "There should be no errors")
	assert.Equal(t, res, []string{"leaf1", "leaf2", "mid1", "leaf3", "mid2"})
}

func TestDepGraphBad(t *testing.T) {
	graph := config.NewDepGraph()

	graph.AddDependency("top", "mid1")
	graph.AddDependency("mid1", "leaf1")
	graph.AddDependency("leaf1", "top")

	res, err := graph.GetDeps("top")

	assert.Nil(t, res)
	assert.EqualError(t, err, "Detected dependency graph cycle at 'top'")

	res, err = graph.GetDeps("bogus")
	assert.Nil(t, res)
	assert.EqualError(t, err, "There is no key 'bogus' in the dependency graph")
}
