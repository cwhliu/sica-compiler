package forge

import "fmt"

type Node struct {
	name  string
	kind  NodeKind
	op    NodeOp
	level int

	fanins  []*Node
	fanouts []*Node
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
func (n *Node) AddFanin(fi *Node) { n.fanins = append(n.fanins, fi) }

/*
Get the number of fanin nodes
*/
func (n *Node) NumFanins() int { return len(n.fanins) }

/*
Get the fanin node at the given index
*/
func (n *Node) Fanin(index int) (*Node, error) {
	if index < 0 || index >= len(n.fanins) {
		return nil, fmt.Errorf("node %s has no fanin[%d]", n.name, index)
	}

	return n.fanins[index], nil
}

/*
Remove node fi from node n's fanin list
*/
func (n *Node) RemoveFanin(fi *Node) {
	for i, nd := range n.fanins {
		if nd == fi {
			n.fanins = append(n.fanins[:i], n.fanins[i+1:]...)
			return
		}
	}
}

/*
Replace old fanin node (oldFi) by new fanin node (newFi)
*/
func (n *Node) ReplaceFanin(oldFi, newFi *Node) {
	for i, nd := range n.fanins {
		if nd == oldFi {
			n.fanins[i] = newFi
			return
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
func (n *Node) Fanout(index int) (*Node, error) {
	if index < 0 || index >= len(n.fanouts) {
		return nil, fmt.Errorf("node %s has no fanout[%d]", n.name, index)
	}

	return n.fanouts[index], nil
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
func (n *Node) ReplaceFanout(oldFo, newFo *Node) {
	for i, nd := range n.fanouts {
		if nd == oldFo {
			n.fanouts[i] = newFo
			return
		}
	}
}
