package forge

import (
	"fmt"
	"strconv"
)

type graph struct {
	allNodes       map[string]*node
	inputNodes     map[string]*node
	outputNodes    map[string]*node
	internalNodes  map[string]*node
	operationNodes map[string]*node
	constantNodes  map[string]*node
}

/*
Create and initialize a graph, and return a pointer to the graph
*/
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

/*
Get the number of total nodes
*/
func (g *graph) numAllNodes() int {
	return len(g.allNodes)
}

/*
Add an operation node to the graph
*/
func (g *graph) addOperationNode(opString string) *node {
	name := "OPR" + strconv.Itoa(len(g.operationNodes))

	if _, exist := nodeOpLUT[opString]; !exist {
		fmt.Println("graph error - unsupported operation", opString)
		return nil
	}

	newNode := &node{name: name, kind: operation, op: nodeOpLUT[opString]}

	g.allNodes[name] = newNode
	g.operationNodes[name] = newNode

	return newNode
}

/*
Get a node by its name, create the node if it doesn't exist
*/
func (g *graph) getNodeByName(name string) *node {
	// Create a new node if a node with the same name does not exist
	if _, exist := g.allNodes[name]; !exist {
		var newNode *node

		switch name[0:3] {
		default:
			fmt.Println("graph error - incorrect node name format")
			return nil
		case "OPR":
			fmt.Println("graph error - should not create operation node here")
			return nil
		case "CON":
			newNode = &node{name: name, kind: constant, op: equal}
			g.constantNodes[name] = newNode
		case "VAR", "ARR":
			// Variable node created here has undetermined node kind because we don't
			// know if it's an input, output, or internal node
			newNode = &node{name: name, op: equal}
		}

		g.allNodes[name] = newNode
	}

	return g.allNodes[name]
}

/*
Things need to be done before the graph can be used
*/
func (g *graph) finalize() {
	// Determine node kind for variable nodes
	for name, node := range g.allNodes {
		if node.kind == undetermined {
			if node.numFanins() == 0 {
				g.inputNodes[name] = node
			} else if node.numFanouts() == 0 {
				g.outputNodes[name] = node
			} else {
				g.internalNodes[name] = node
			}
		}
	}
}
