package forge

import (
	"fmt"
	"math"
)

type Node struct {
	name  string
	kind  NodeKind
	op    NodeOp
	level int

	fanins  []*Node
	fanouts []*Node

	faninSigns []bool

	value float64
}

// -----------------------------------------------------------------------------

/*
Node n receives node fi as a fanin
*/
func (n *Node) Receive(fi *Node) {
	// Set fi to be n's fanin
	n.AddFanin(fi)
	// Set n to be fi's fanout
	fi.AddFanout(n)
}

// Fanins
// -----------------------------------------------------------------------------

/*
Add node fi to node n's fanin list
*/
func (n *Node) AddFanin(fi *Node) {
	n.fanins = append(n.fanins, fi)
	n.faninSigns = append(n.faninSigns, false)
}

/*
Get the number of fanin nodes
*/
func (n *Node) NumFanins() int { return len(n.fanins) }

/*
Get the fanin node at the given index
*/
func (n *Node) Fanin(index int) *Node {
	return n.fanins[index]
}

/*
Remove node fi from node n's fanin list
*/
func (n *Node) RemoveFanin(fi *Node) {
	for i, nd := range n.fanins {
		if nd == fi {
			n.fanins = append(n.fanins[:i], n.fanins[i+1:]...)
			n.faninSigns = append(n.faninSigns[:i], n.faninSigns[i+1:]...)
			return
		}
	}
}

/*
Replace old fanin node (oldFi) by new fanin node (newFi)
*/
func (n *Node) ReplaceFanin(oldFi, newFi *Node) int {
	for i, nd := range n.fanins {
		if nd == oldFi {
			n.fanins[i] = newFi
			return i
		}
	}
	return -1
}

func (n *Node) GetFaninSignByIndex(index int) bool { return n.faninSigns[index] }

func (n *Node) GetFaninSignByNode(fi *Node) bool {
	for i, nd := range n.fanins {
		if nd == fi {
			return n.faninSigns[i]
		}
	}
	return false // this actually should be an error
}

func (n *Node) NegateFaninByNode(fi *Node) {
	for i, nd := range n.fanins {
		if nd == fi {
			n.faninSigns[i] = !n.faninSigns[i]
			return
		}
	}
}

func (n *Node) NegateFaninByIndex(index int) {
	n.faninSigns[index] = !n.faninSigns[index]
}

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

// Fanouts
// -----------------------------------------------------------------------------

/*
Add node fo to node n's fanout list
*/
func (n *Node) AddFanout(fo *Node) { n.fanouts = append(n.fanouts, fo) }

/*
Get the number of fanout nodes
*/
func (n *Node) NumFanouts() int { return len(n.fanouts) }

/*
Get the fanout node at the given index
*/
func (n *Node) Fanout(index int) *Node {
	return n.fanouts[index]
}

/*
Remove node fo from node n's fanout list
*/
func (n *Node) RemoveFanout(fo *Node) {
	for i, nd := range n.fanouts {
		if nd == fo {
			n.fanouts = append(n.fanouts[:i], n.fanouts[i+1:]...)
			return
		}
	}
}

/*
Replace old fanout node (oldFo) by new fanout node (newFo)
*/
func (n *Node) ReplaceFanout(oldFo, newFo *Node) int {
	for i, nd := range n.fanouts {
		if nd == oldFo {
			n.fanouts[i] = newFo
			return i
		}
	}
	return -1
}

// Node value evaluation
// -----------------------------------------------------------------------------
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
