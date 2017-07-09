package forge

import "fmt"

type graph struct {
	allNodes       map[string]*node
	inputNodes     map[string]*node
	outputNodes    map[string]*node
	internalNodes  map[string]*node
	operationNodes map[string]*node
	constantNodes  map[string]*node
}

// Create and initialize a graph, and return a pointer to the graph
func createGraph() *graph {
	g := &graph{}

	g.allNodes = make(map[string]*node)
	g.inputNodes = make(map[string]*node)
	g.outputNodes = make(map[string]*node)
	g.internalNodes = make(map[string]*node)
	g.operationNodes = make(map[string]*node)
	g.constantNodes = make(map[string]*node)

	return g
}

func (g *graph) numAllNodes() int {
	return len(g.allNodes)
}

// Add an internal node to the graph
func (g *graph) addInternalNode(name string) bool {
	// Check if the name is duplicated
	if _, exist := g.allNodes[name]; exist {
		fmt.Printf("graph error - adding duplicated internal node %s\n", name)

		return false
	}

	n := createNode(name, internal, equal)

	g.allNodes[name] = n
	g.internalNodes[name] = n

	return true
}
