package forge

import "fmt"

type Node struct {
	name string
	kind NodeKind
	op   NodeOp

	fanins  []*Node
	fanouts []*Node
}

// -----------------------------------------------------------------------------

/*
Node n receives node fi as a fanin
*/
func (n *Node) receive(fi *Node) {
	// Set fi to be n's fanin
	n.fanins = append(n.fanins, fi)
	// Set n to be fi's fanout
	fi.fanouts = append(fi.fanouts, n)
}

// Fanins
// -----------------------------------------------------------------------------

/*
Add one node to the fanin list
*/
func (n *Node) addFanin(fi *Node) {
	n.fanins = append(n.fanins, fi)
}

/*
Get the number of fanin nodes
*/
func (n *Node) numFanins() int {
	return len(n.fanins)
}

/*
Get the fanin node at the given index
*/
func (n *Node) fanin(index int) (*Node, error) {
	if index < 0 || index >= len(n.fanins) {
		return nil, fmt.Errorf("node %s has no fanin[%d]", n.name, index)
	}

	return n.fanins[index], nil
}

func (n *Node) removeFanin(fi *Node) {
  for i, nd := range n.fanins {
    if nd == fi {
      n.fanins = append(n.fanins[:i], n.fanins[i+1:]...)
      return
    }
  }
}

func (n *Node) replaceFanin(oldFi, newFi *Node) {
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
Add one node to the fanout list
*/
func (n *Node) addFanout(fo *Node) {
	n.fanouts = append(n.fanouts, fo)
}

/*
Get the number of fanout nodes
*/
func (n *Node) numFanouts() int {
	return len(n.fanouts)
}

/*
Get the fanout node at the given index
*/
func (n *Node) fanout(index int) (*Node, error) {
	if index < 0 || index >= len(n.fanouts) {
		return nil, fmt.Errorf("node %s has no fanout[%d]", n.name, index)
	}

	return n.fanouts[index], nil
}
func (n *Node) removeFanout(fo *Node) {
  for i, nd := range n.fanouts {
    if nd == fo {
      n.fanouts = append(n.fanouts[:i], n.fanouts[i+1:]...)
      return
    }
  }
}

func (n *Node) replaceFanout(oldFo, newFo *Node) {
  for i, nd := range n.fanouts {
    if nd == oldFo {
      n.fanouts[i] = newFo
      return
    }
  }
}

