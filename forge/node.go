package forge

import "fmt"

type node struct {
	name string
	kind nodeKind
	op   nodeOp

	fanins  []*node
	fanouts []*node
}

// -----------------------------------------------------------------------------

/*
Node n receives node nd as a fanin
*/
func (n *node) receive(nd *node) {
	// Set nd to be n's fanin
	n.fanins = append(n.fanins, nd)
	// Set n to be nd's fanout
	nd.fanouts = append(nd.fanouts, n)
}

// Fanins
// -----------------------------------------------------------------------------

/*
Add one node to the fanin list
*/
func (n *node) addFanin(nd *node) {
	n.fanins = append(n.fanins, nd)
}

/*
Get the number of fanin nodes
*/
func (n *node) numFanins() int {
	return len(n.fanins)
}

/*
Get the fanin node at the given index
*/
func (n *node) fanin(index int) (*node, error) {
	if index < 0 || index >= len(n.fanins) {
		return nil, fmt.Errorf("node %s has no fanin[%d]", n.name, index)
	}

	return n.fanins[index], nil
}

// Fanouts
// -----------------------------------------------------------------------------

/*
Add one node to the fanout list
*/
func (n *node) addFanout(nd *node) {
	n.fanouts = append(n.fanouts, nd)
}

/*
Get the number of fanout nodes
*/
func (n *node) numFanouts() int {
	return len(n.fanouts)
}

/*
Get the fanout node at the given index
*/
func (n *node) fanout(index int) (*node, error) {
	if index < 0 || index >= len(n.fanouts) {
		return nil, fmt.Errorf("node %s has no fanout[%d]", n.name, index)
	}

	return n.fanouts[index], nil
}
