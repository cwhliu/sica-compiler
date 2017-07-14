package forge

import "container/heap"

/*
A wrapper around the underlying heap to provide simple and clean methods for
the priority queue operation
*/
type NodePQ struct {
	pq *nodePQ
}

func CreateNodePQ() *NodePQ {
	PQ := &NodePQ{}

	PQ.pq = &nodePQ{}
	heap.Init(PQ.pq)

	return PQ
}

func (PQ *NodePQ) Len() int { return PQ.pq.Len() }

func (PQ *NodePQ) Push(n NodePQEntry) { heap.Push(PQ.pq, n) }

func (PQ *NodePQ) PopMin() *Node {
	return heap.Pop(PQ.pq).(NodePQEntry).node
}

func (PQ *NodePQ) PopMax() *Node {
	return heap.Remove(PQ.pq, PQ.pq.Len()-1).(NodePQEntry).node
}

func (PQ *NodePQ) GetNodeByIndex(index int) *Node {
	return (*PQ.pq)[index].node
}

func (PQ *NodePQ) FindNode(node *Node) bool {
	for _, entry := range *PQ.pq {
		if entry.node == node {
			return true
		}
	}

	return false
}

// Underlying heap container
//  Code adapted from Go's heap package documentation
// -----------------------------------------------------------------------------

type NodePQEntry struct {
	node     *Node
	priority int
}

type nodePQ []NodePQEntry

func (pq nodePQ) Len() int { return len(pq) }

func (pq nodePQ) Less(i, j int) bool { return pq[i].priority < pq[j].priority }

func (pq nodePQ) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }

func (pq *nodePQ) Push(x interface{}) {
	entry := x.(NodePQEntry)
	*pq = append(*pq, entry)
}

func (pq *nodePQ) Pop() interface{} {
	oldPQ := *pq
	entry := oldPQ[len(oldPQ)-1]
	*pq = oldPQ[0 : len(oldPQ)-1]
	return entry
}
