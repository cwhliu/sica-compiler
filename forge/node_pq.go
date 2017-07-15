package forge

import "container/heap"

/*
NodePQ is a wrapper around the underlying heap container, providing simple and
clean methods for priority queue operations.
*/
type NodePQ struct {
	pq *nodePQ
}

/*
NodePQEntry is the storage entry of NodePQ.
*/
type NodePQEntry struct {
	Payload  *Node
	Priority int
}

// -----------------------------------------------------------------------------

/*
CreateNodePQ creates and returns a priority queue of nodes.
*/
func CreateNodePQ() *NodePQ {
	PQ := &NodePQ{}

	PQ.pq = &nodePQ{}
	heap.Init(PQ.pq)

	return PQ
}

/*
Len returns the length of the priority queue.
*/
func (PQ *NodePQ) Len() int { return PQ.pq.Len() }

/*
Push pushes a new entry to the priority queue.
*/
func (PQ *NodePQ) Push(n NodePQEntry) { heap.Push(PQ.pq, n) }

/*
PopMin pops an entry with the minimum priority from the priority queue.
*/
func (PQ *NodePQ) PopMin() *Node {
	return heap.Pop(PQ.pq).(NodePQEntry).Payload
}

/*
PopMax pops an entry with the maximum priority from the priority queue.
*/
func (PQ *NodePQ) PopMax() *Node {
	return heap.Remove(PQ.pq, PQ.pq.Len()-1).(NodePQEntry).Payload
}

/*
GetNodeByIndex returns the node stored in the priority queue at index.
*/
func (PQ *NodePQ) GetNodeByIndex(index int) *Node {
	return (*PQ.pq)[index].Payload
}

/*
FindNode finds if a node exists in the priority queue.
*/
func (PQ *NodePQ) FindNode(node *Node) bool {
	for _, entry := range *PQ.pq {
		if entry.Payload == node {
			return true
		}
	}
	return false
}

// Underlying heap container (code adapted from Go's heap package documentation)
// -----------------------------------------------------------------------------

type nodePQ []NodePQEntry

func (pq nodePQ) Len() int { return len(pq) }

func (pq nodePQ) Less(i, j int) bool { return pq[i].Priority < pq[j].Priority }

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
