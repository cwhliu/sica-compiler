package forge

import (
	"fmt"
	"strconv"
)

/*
Graph represents the graph structure built by the parser based on a C++ source file.
*/
type Graph struct {
	allNodes       map[string]*Node
	inputNodes     map[string]*Node
	outputNodes    map[string]*Node
	operationNodes map[string]*Node
	constantNodes  map[string]*Node

	inputValues  []map[string]float64
	outputValues []map[string]float64
}

// -----------------------------------------------------------------------------

/*
CreateGraph creates and returns a pointer to an initialized graph.
*/
func CreateGraph() *Graph {
	g := &Graph{}

	g.allNodes = make(map[string]*Node)
	g.inputNodes = make(map[string]*Node)
	g.outputNodes = make(map[string]*Node)
	g.operationNodes = make(map[string]*Node)
	g.constantNodes = make(map[string]*Node)

	return g
}

/*
Legalize fills in some missing information after the graph is built.

This function should be invoked after the graph is built.
*/
func (g *Graph) Legalize() {
	// Determine node kind for undetermined nodes and delete internal nodes
	for name, node := range g.allNodes {
		if node.kind == NodeKind_Undetermined {
			if node.NumFanins() == 0 {
				// An undetermined node without fanin is an input to the graph
				node.kind = NodeKind_Input
				g.inputNodes[name] = node
			} else if node.NumFanouts() == 0 {
				// An undetermined node without fanout is an output of the graph
				node.kind = NodeKind_Output
				g.outputNodes[name] = node
			} else {
				// Otherwise it's an internal node created in the source file
				// We don't need them so delete these nodes here
				fi := node.Fanin(0)
				fi.RemoveFanout(node)

				for _, fo := range node.fanouts {
					fi.AddFanout(fo)
					fo.ReplaceFanin(node, fi)
				}

				g.DeleteNodeByName(name)
			}
		}
	}

	// Do some simple operation transformation here
	g.Levelize()

	pq := CreateNodePQ()
	for _, node := range g.operationNodes {
		pq.Push(NodePQEntry{node, node.level})
	}

	for pq.Len() > 0 {
		node := pq.PopMin()

		switch node.op {
		case NodeOp_Sub:
			if node.NumFanins() == 1 {
				// A subtraction with one fanin is a negate operation
				// Delete the operation node, pass its fanin and sign to fanouts
				fi := node.Fanin(0)
				fi.RemoveFanout(node)

				for _, fo := range node.fanouts {
					fi.AddFanout(fo)
					fo.ReplaceFanin(node, fi)
					fo.NegateFaninByNode(fi)
				}

				g.DeleteNodeByName(node.name)
			} else {
				// Change a subtraction into addition by negating its second fanin
				node.op = NodeOp_Add
				node.NegateFaninByNode(node.Fanin(1))
			}
		}

		node.PropagateSign()
	}

	// Set the value for constant nodes
	for _, n := range g.constantNodes {
		n.value, _ = strconv.ParseFloat(n.name[3:], 64)
	}
}

// -----------------------------------------------------------------------------

func (g *Graph) NumAllNodes() int       { return len(g.allNodes) }
func (g *Graph) NumInputNodes() int     { return len(g.inputNodes) }
func (g *Graph) NumOutputNodes() int    { return len(g.outputNodes) }
func (g *Graph) NumOperationNodes() int { return len(g.operationNodes) }
func (g *Graph) NumConstantNodes() int  { return len(g.constantNodes) }

// -----------------------------------------------------------------------------

/*
AddOperationNode adds an operation node to the graph.

Other kinds of nodes should be created by GetNodeByName().
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
GetNodeByName get a node by the name, create the node if it doesn't exist.

Operation nodes should be created by AddOperationNode().
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
DeleteNodeByName deletes a node by the name from the graph.
*/
func (g *Graph) DeleteNodeByName(name string) {
	switch g.allNodes[name].kind {
	case NodeKind_Input:
		delete(g.inputNodes, name)
	case NodeKind_Output:
		delete(g.outputNodes, name)
	case NodeKind_Operation:
		delete(g.operationNodes, name)
	case NodeKind_Constant:
		delete(g.constantNodes, name)
	}

	delete(g.allNodes, name)
}

// -----------------------------------------------------------------------------

/*
Levelize calculates the level of each node recursively. Input and constant nodes
are at level 0.
*/
func (g *Graph) Levelize() int {
	var maxLevel int = -10

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

			if n.level > maxLevel {
				maxLevel = n.level
			}
		}
	}

	for _, node := range g.outputNodes {
		levelize(node)
	}

	return maxLevel
}

// -----------------------------------------------------------------------------

/*
Eval evaluates all nodes' value. The graph is first levelized and then nodes are
evaluated starting from the lowest level.
*/
func (g *Graph) Eval() {
	g.Levelize()

	pq := CreateNodePQ()
	for _, n := range g.operationNodes {
		pq.Push(NodePQEntry{n, n.level})
	}
	for _, n := range g.outputNodes {
		pq.Push(NodePQEntry{n, n.level})
	}

	for pq.Len() > 0 {
		node := pq.PopMin()
		node.Eval()
	}
}
