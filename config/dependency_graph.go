package config // FIXME: Perhaps should be in its own package

import (
	"fmt"

	orderedmap "github.com/wk8/go-ordered-map"
)

// Node represents a node in a dependency graph
type Node struct {
	key string
	// Maps dependency name to 'true'
	dependencies *orderedmap.OrderedMap
}

// DepGraph represents a dependency graph
type DepGraph struct {
	nodes map[string]*Node
}

// NewDepGraph returns a new DepGraph
func NewDepGraph() *DepGraph {
	return &DepGraph{nodes: make(map[string]*Node)}
}

// EnsureNode ensures that there is a node in DepGraph with the specified key
func (graph *DepGraph) EnsureNode(key string) *Node {
	n := graph.nodes[key]

	if n == nil {
		// Prepare a fresh node and add it to the graph
		n = &Node{key: key, dependencies: orderedmap.New()}
		graph.nodes[key] = n
	}

	return n
}

// AddDependency adds a record indicating that 'key' depends on 'dep' (both strings).
// The order of dependencies is preserved and duplicates are excluded.
// Cycles are not detected at add time but they will be detected by GetDeps
func (graph *DepGraph) AddDependency(key string, dep string) {
	n := graph.EnsureNode(key)

	n.dependencies.Set(dep, true)

	// Also ensure that a node exists for the dep
	graph.EnsureNode(dep)
}

// Unexported helper for use by GetDeps
func (graph *DepGraph) getDeps(key string, processing map[string]bool) ([]string, error) {
	n := graph.nodes[key]

	if n == nil {
		return nil, fmt.Errorf("There is no key '%s' in the dependency graph", key)
	}

	if n.dependencies.Len() == 0 {
		// Base case
		return nil, nil
	}

	// This node has dependencies which need to be recursively examined.
	// Since we're about to recurse, first check that we haven't already been at this
	// node, which indicates a cycle.
	if processing[key] {
		return nil, fmt.Errorf("Detected dependency graph cycle at '%s'", key)
	}

	processing[key] = true
	defer func() { processing[key] = false }()

	res := orderedmap.New()

	for pair := n.dependencies.Oldest(); pair != nil; pair = pair.Next() {
		dep := pair.Key.(string)
		childDeps, err := graph.getDeps(dep, processing)

		if err != nil {
			return nil, err
		}

		for _, childDep := range childDeps {
			res.Set(childDep, true)
		}

		res.Set(dep, true)
	}

	return orderedMapToList(res), nil
}

func orderedMapToList(om *orderedmap.OrderedMap) []string {
	res := []string{}

	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		res = append(res, pair.Key.(string))
	}

	return res
}

// GetDeps returns a slice of strings representing the direct and indirect
// dependencies of 'key'.  The dependencies will be returned in the order that
// they should be processed (i.e. leaves first)
func (graph *DepGraph) GetDeps(key string) ([]string, error) {
	return graph.getDeps(key, make(map[string]bool))
}
