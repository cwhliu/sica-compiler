package forge

import "fmt"

type node struct {
	name string
	kind nodeKind
	op   nodeOp

	fanins  []*node
	fanouts []*node
}

// Create and initialize a node, and return a pointer to the node
func createNode(name string, kind nodeKind, op nodeOp) *node {
	n := &node{name: name, kind: kind, op: op}

	return n
}

// Fanins
// -----------------------------------------------------------------------------

// Add one node to the fanin list
func (n *node) addFanin(newNode *node) {
	n.fanins = append(n.fanins, newNode)
}

// Get the number of fanin nodes
func (n *node) numFanins() int {
	return len(n.fanins)
}

// Get the fanin node at the given index
func (n *node) fanin(index int) (*node, error) {
	if index < 0 || index >= len(n.fanins) {
		return nil, fmt.Errorf("node %s has no fanin[%d]", n.name, index)
	}

	return n.fanins[index], nil
}

// Fanouts
// -----------------------------------------------------------------------------

// Add one node to the fanout list
func (n *node) addFanout(newNode *node) {
	n.fanouts = append(n.fanouts, newNode)
}

// Get the number of fanout nodes
func (n *node) numFanouts() int {
	return len(n.fanouts)
}

// Get the fanout node at the given index
func (n *node) fanout(index int) (*node, error) {
	if index < 0 || index >= len(n.fanouts) {
		return nil, fmt.Errorf("node %s has no fanout[%d]", n.name, index)
	}

	return n.fanouts[index], nil
}
