package forge

import "fmt"

/*
Delete all internal nodes created in the source file to hold temporary results
*/
func (g *Graph) OptimizeInternalNodes() {
	for name, node := range g.internalNodes {
		if node.NumFanins() != 1 {
			fmt.Println("optimizer error - internal nodes should have single fanin")
			return
		}

		fi := node.Fanin(0)
		fi.RemoveFanout(node)

		for _, fo := range node.fanouts {
			fi.AddFanout(fo)
			fo.ReplaceFanin(node, fi)
		}

		g.DeleteNodeByName(name)
	}
}

/*
Eliminate duplicated operations using value numbering

The graph must be levelized and value numbering needs to start from inputs,
otherwise a duplicated operation will not be eliminated when its fanout is
processed before it (this happens if loop over a map because map is not ordered)

See "Engineering a Compiler 2nd Edition, section 8.4.1"
*/
func (g *Graph) OptimizeValueNumbering() {
	// Levelize the graph
	g.Levelize()

	// Use a priority queue to sort operation nodes by level
	pq := CreateNodePQ()
	for _, node := range g.operationNodes {
		pq.Push(NodePQEntry{node, node.level})
	}

	// This map acts as a hash holding value numbers
	vnMap := make(map[string]*Node)

	for pq.Len() > 1 {
		node := pq.PopMin()

		// Construct the value number for this operation
		// Here we're not using fanin's value number but their name, this is
		// sub-optimal but much easier
		vnKey := NodeOpStringLUT[node.op]
		for _, fi := range node.fanins {
			vnKey += fi.name
		}

		if vnNode, exist := vnMap[vnKey]; !exist {
			// Store the operation if it does not exist
			vnMap[vnKey] = node
		} else {
			// Otherwise replace the operation with the existing one

			for _, fi := range node.fanins {
				fi.RemoveFanout(node)
			}

			for _, fo := range node.fanouts {
				vnNode.AddFanout(fo)
				fo.ReplaceFanin(node, vnNode)
			}

			g.DeleteNodeByName(node.name)
		}
	}
}

func (g *Graph) OptimizeTreeHeight() {
	roots := CreateNodePQ()
	ranks := make(map[*Node]int)

	for _, node := range g.allNodes {
		ranks[node] = -1
	}

	for _, node := range g.operationNodes {
		if node.op == NodeOp_Add {
			if node.NumFanouts() > 1 ||
				node.NumFanouts() == 1 && node.op != node.Fanout(0).op {
				roots.Push(NodePQEntry{node, 1})
			}
		}
	}

	var balance func(n *Node)
	balance = func(n *Node) {
		if ranks[n] >= 0 {
			return
		}

		q := CreateNodePQ()
		r := []*Node{}

		var flatten func(n *Node, op NodeOp) int
		var rebuild func(n *Node)

		flatten = func(n *Node, op NodeOp) int {
			if n.kind == NodeKind_Constant {
				ranks[n] = 0
				q.Push(NodePQEntry{n, ranks[n]})
			} else if n.kind == NodeKind_Input || n.op != op {
				ranks[n] = 1
				q.Push(NodePQEntry{n, ranks[n]})
			} else if exist := roots.FindNode(n); exist {
				balance(n)
				q.Push(NodePQEntry{n, ranks[n]})
			} else {
				ranks[n] = flatten(n.Fanin(0), n.op) + flatten(n.Fanin(1), n.op)
				r = append(r, n)
			}

			return ranks[n]
		}

		rebuild = func(n *Node) {
			if q.Len() == 2 {
				return
			}

			for n.NumFanins() > 0 {
				n.Fanin(0).RemoveFanout(n)
				n.RemoveFanin(n.Fanin(0))
			}

			for _, node := range r {
				for node.NumFanins() > 0 {
					node.Fanin(0).RemoveFanout(node)
					node.RemoveFanin(node.Fanin(0))
				}
				for node.NumFanouts() > 0 {
					node.Fanout(0).RemoveFanin(node)
					node.RemoveFanout(node.Fanout(0))
				}
			}

			for q.Len() > 0 {
				var nodeL *Node = q.PopMin()
				var nodeR *Node = q.PopMin()
				var nodeT *Node

				if q.Len() == 0 {
					nodeT = n
				} else {
					nodeT, r = r[len(r)-1], r[:len(r)-1]
				}

				nodeL.AddFanout(nodeT)
				nodeR.AddFanout(nodeT)

				nodeT.AddFanin(nodeL)
				nodeT.AddFanin(nodeR)

				ranks[nodeT] = ranks[nodeL] + ranks[nodeR]

				if q.Len() != 0 {
					q.Push(NodePQEntry{nodeT, ranks[nodeT]})
				}
			}
		}

		ranks[n] = flatten(n.Fanin(0), n.op) + flatten(n.Fanin(1), n.op)

		rebuild(n)
	}

	for roots.Len() > 0 {
		balance(roots.PopMin())
	}
}
