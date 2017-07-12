package forge

import (
	"fmt"
	"strconv"
)

type Graph struct {
	allNodes       map[string]*Node
	inputNodes     map[string]*Node
	outputNodes    map[string]*Node
	internalNodes  map[string]*Node
	operationNodes map[string]*Node
	constantNodes  map[string]*Node
}

// -----------------------------------------------------------------------------

/*
Create and initialize a graph, and return a pointer to the graph
*/
func CreateGraph() *Graph {
	g := &Graph{}

	g.allNodes = make(map[string]*Node)
	g.inputNodes = make(map[string]*Node)
	g.outputNodes = make(map[string]*Node)
	g.internalNodes = make(map[string]*Node)
	g.operationNodes = make(map[string]*Node)
	g.constantNodes = make(map[string]*Node)

	return g
}

/*
Update undetermined node types after the graph is built
*/
func (g *Graph) UpdateNodeType() {
	// Determine node kind for variable nodes
	for name, node := range g.allNodes {
		if node.kind == NodeKind_Undetermined {
			if node.NumFanins() == 0 {
				g.inputNodes[name] = node
			} else if node.NumFanouts() == 0 {
				g.outputNodes[name] = node
			} else {
				g.internalNodes[name] = node
			}
		}
	}
}

// -----------------------------------------------------------------------------

/*
Get the number of total nodes
*/
func (g *Graph) NumAllNodes() int {
	return len(g.allNodes)
}

// -----------------------------------------------------------------------------

/*
Add an operation node to the graph
*/
func (g *Graph) AddOperationNode(opString string) *Node {
	name := "OPR" + strconv.Itoa(len(g.operationNodes))

	if _, exist := NodeOpLUT[opString]; !exist {
		fmt.Println("graph error - unsupported operation", opString)
		return nil
	}

	newNode := &Node{name: name, kind: NodeKind_Operation, op: NodeOpLUT[opString]}

	g.allNodes[name] = newNode
	g.operationNodes[name] = newNode

	return newNode
}

/*
Get a node by its name, create the node if it doesn't exist
*/
func (g *Graph) GetNodeByName(name string) *Node {
	// Create a new node if a node with the same name does not exist
	if _, exist := g.allNodes[name]; !exist {
		var newNode *Node

		switch name[0:3] {
		default:
			fmt.Println("graph error - incorrect node name format")
			return nil
		case "OPR":
			fmt.Println("graph error - should not create operation node here")
			return nil
		case "CON":
			newNode = &Node{name: name, kind: NodeKind_Constant, op: NodeOp_Equal}
			g.constantNodes[name] = newNode
		case "VAR", "ARR":
			// Variable node created here has undetermined node kind because we don't
			// know if it's an input, output, or internal node
			newNode = &Node{name: name, op: NodeOp_Equal}
		}

		g.allNodes[name] = newNode
	}

	return g.allNodes[name]
}

/*
Delete a node by its name
*/
func (g *Graph) DeleteNodeByName(name string) {
	switch g.allNodes[name].kind {
	case NodeKind_Input:
		delete(g.inputNodes, name)
	case NodeKind_Output:
		delete(g.outputNodes, name)
	case NodeKind_Internal:
		delete(g.internalNodes, name)
	case NodeKind_Operation:
		delete(g.operationNodes, name)
	case NodeKind_Constant:
		delete(g.constantNodes, name)
	}

	delete(g.allNodes, name)
}

// -----------------------------------------------------------------------------

func (g *Graph) Levelize() {
	for _, node := range g.allNodes {
		node.level = -1
	}

	var levelize func(n *Node)
	levelize = func(n *Node) {
		if n.kind == NodeKind_Input || n.kind == NodeKind_Constant {
			n.level = 0
		} else {
			max := -1000

			for _, fi := range n.fanins {
				levelize(fi)

				if fi.level > max {
					max = fi.level
				}
			}

			n.level = max + 1
		}
	}

	for _, node := range g.outputNodes {
		levelize(node)
	}
}
