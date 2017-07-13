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

		fi, _ := node.Fanin(0)
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
}
