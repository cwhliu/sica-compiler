package forge

import (
	"fmt"
	"math"
)

/*
Node is the basic unit in a graph.
*/
type Node struct {
	name string
	kind NodeKind
	op   NodeOp

	level int

	fanins  []*Node
	fanouts []*Node

	faninSigns []bool

	value float64

	actualStartTime   int
	actualFinishTime  int
	processorAssigned int

	isScheduled bool
	pgScheduled int
	peScheduled int
	startTime   int
	finishTime  int
}

/*
CreateNode creates and returns a pointer to an initialized node.
*/
func CreateNode(name string, kind NodeKind, op NodeOp) *Node {
	node := &Node{}

	node.name = name
	node.kind = kind
	node.op = op

	node.isScheduled = false
	node.pgScheduled = -1
	node.peScheduled = -1
	node.startTime = -1
	node.finishTime = -1

	return node
}

// Fanin
// -----------------------------------------------------------------------------

/*
AddFanin connects fi as a fanin to the node.
*/
func (n *Node) AddFanin(fi *Node) {
	n.fanins = append(n.fanins, fi)
	n.faninSigns = append(n.faninSigns, false)
}

/*
NumFanins returns the number of fanins to the node.
*/
func (n *Node) NumFanins() int { return len(n.fanins) }

/*
Fanin returns the fanin at the index.
*/
func (n *Node) Fanin(index int) *Node {
	return n.fanins[index]
}

/*
RemoveFanin disconnects fi from the node's fanin.
*/
func (n *Node) RemoveFanin(fi *Node) {
	for i, node := range n.fanins {
		if node == fi {
			n.fanins = append(n.fanins[:i], n.fanins[i+1:]...)
			n.faninSigns = append(n.faninSigns[:i], n.faninSigns[i+1:]...)
			return
		}
	}
}

/*
ReplaceFanin replaces an old fanin by a new fanin.
*/
func (n *Node) ReplaceFanin(oldFi, newFi *Node) int {
	for i, node := range n.fanins {
		if node == oldFi {
			n.fanins[i] = newFi
			return i
		}
	}
	return -1
}

// Fanin sign
// -----------------------------------------------------------------------------

/*
GetFaninSignByIndex returns the sign for the fanin at the index.
*/
func (n *Node) GetFaninSignByIndex(index int) bool { return n.faninSigns[index] }

/*
GetFaninSignByNode return the sign for fi.
*/
func (n *Node) GetFaninSignByNode(fi *Node) bool {
	for i, node := range n.fanins {
		if node == fi {
			return n.faninSigns[i]
		}
	}
	return false // this actually should be an error
}

/*
NegateFaninByIndex negates the sign for the fanin at the index.
*/
func (n *Node) NegateFaninByIndex(index int) {
	n.faninSigns[index] = !n.faninSigns[index]
}

/*
NegateFaninByNode negates the sign for fi.
*/
func (n *Node) NegateFaninByNode(fi *Node) {
	for i, node := range n.fanins {
		if node == fi {
			n.faninSigns[i] = !n.faninSigns[i]
			return
		}
	}
}

/*
PropagateSign propagates fanin signs to the fanout.
Negates the fanout of an addition if both fanins are negative.
Negates the fanout of a multiply and division if one of the fanins is negative.
*/
func (n *Node) PropagateSign() {
	switch n.op {
	case NodeOp_Add:
		if n.faninSigns[0] && n.faninSigns[1] {
			n.faninSigns[0], n.faninSigns[1] = false, false

			for _, fo := range n.fanouts {
				fo.NegateFaninByNode(n)
			}
		}
	case NodeOp_Mul, NodeOp_Div:
		if n.faninSigns[0] && !n.faninSigns[1] ||
			!n.faninSigns[0] && n.faninSigns[1] {
			n.faninSigns[0], n.faninSigns[1] = false, false

			for _, fo := range n.fanouts {
				fo.NegateFaninByNode(n)
			}
		} else if !n.faninSigns[0] && !n.faninSigns[1] {
			n.faninSigns[0], n.faninSigns[1] = false, false
		}
	}
}

// Fanout
// -----------------------------------------------------------------------------

/*
AddFanout connects fo as a fanout of the node.
*/
func (n *Node) AddFanout(fo *Node) { n.fanouts = append(n.fanouts, fo) }

/*
NumFanouts returns the number of fanouts of the node
*/
func (n *Node) NumFanouts() int { return len(n.fanouts) }

/*
Fanout returns the fanout at the index.
*/
func (n *Node) Fanout(index int) *Node {
	return n.fanouts[index]
}

/*
RemoveFanout removes fo from the node's fanout.
*/
func (n *Node) RemoveFanout(fo *Node) {
	for i, node := range n.fanouts {
		if node == fo {
			n.fanouts = append(n.fanouts[:i], n.fanouts[i+1:]...)
			return
		}
	}
}

/*
ReplaceFanout replaces an old fanout by a new fanout.
*/
func (n *Node) ReplaceFanout(oldFo, newFo *Node) int {
	for i, node := range n.fanouts {
		if node == oldFo {
			n.fanouts[i] = newFo
			return i
		}
	}
	return -1
}

// -----------------------------------------------------------------------------

/*
Receive connects fi as a fanin to the node, and the node as a fanout of fi.
*/
func (n *Node) Receive(fi *Node) {
	n.AddFanin(fi)
	fi.AddFanout(n)
}

// -----------------------------------------------------------------------------

// Eval evaluates the node's value based on its operation and fanins' value
func (n *Node) Eval() {
	signs := []float64{}
	for _, sign := range n.faninSigns {
		if sign {
			signs = append(signs, -1)
		} else {
			signs = append(signs, 1)
		}
	}

	switch n.op {
	case NodeOp_Equal:
		n.value = n.Fanin(0).value
	case NodeOp_Add:
		n.value = (signs[0] * n.Fanin(0).value) + (signs[1] * n.Fanin(1).value)
	case NodeOp_Sub:
		fmt.Println("node eval error - should not contain subtraction")
	case NodeOp_Mul:
		n.value = (signs[0] * n.Fanin(0).value) * (signs[1] * n.Fanin(1).value)
	case NodeOp_Div:
		n.value = (signs[0] * n.Fanin(0).value) / (signs[1] * n.Fanin(1).value)
	case NodeOp_Power:
		n.value = math.Pow(signs[0]*n.Fanin(0).value, signs[1]*n.Fanin(1).value)
	case NodeOp_Sin:
		n.value = math.Sin(signs[0] * n.Fanin(0).value)
	case NodeOp_Cos:
		n.value = math.Cos(signs[0] * n.Fanin(0).value)
	default:
		fmt.Println("node eval error - unsupported operation", NodeOpStringLUT[n.op])
	}
}
