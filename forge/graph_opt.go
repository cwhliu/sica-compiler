package forge

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

	for pq.Len() > 0 {
		node := pq.PopMin()

		// Construct the value number for this operation
		// Here we're not using fanin's value number but their name, this is
		// sub-optimal but much easier
		vnKey := NodeOpStringLUT[node.op]
		for i, fi := range node.fanins {
			if node.FaninSign(i) {
				vnKey += "-"
			}
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

/*
Optimize tree height in the graph by balancing a skewed trees in two phases
 The first phase identifies candidate tree roots
 The second phase finds all the operands for a candidate tree and build a
 balanced tree for them

See "Engineering a Compiler 2nd Edition, section 8.4.2"
*/
func (g *Graph) OptimizeTreeHeight() {

	// Phase 1 - analysis
	// ---------------------------------------------------------------------------

	// Sort candidate roots by their operation precedence
	candidateRoots := CreateNodePQ()

	// Store nodes' rank, -1 means not processed yet
	// A node's rank is used to build a balanced tree (approximately)
	ranks := make(map[*Node]int)
	for _, node := range g.allNodes {
		ranks[node] = -1
	}

	// Find candidate tree roots
	for _, node := range g.operationNodes {
		// A node is a candidate root if it has multiple fanouts or it's fanout has
		// different operation than its own
		if node.NumFanouts() > 1 ||
			node.NumFanouts() == 1 && node.op != node.Fanout(0).op {
			if node.op == NodeOp_Add {
				candidateRoots.Push(NodePQEntry{node, 1}) // let add have precedence of 1
			}
		}
	}

	// Phase 2 - transformation
	// ---------------------------------------------------------------------------

	var balance func(n *Node)

	// Defer invocations of the function balance to the end when it's defined
	defer func() {
		for candidateRoots.Len() > 0 {
			balance(candidateRoots.PopMin())
		}
	}()

	/*
		This function finds all the operands for a root node (by recursively calling
		flatten) and build a balanced tree for them (by calling rebuild)
	*/
	balance = func(root *Node) {
		if ranks[root] >= 0 { // this tree is already processed
			return
		}

		// Store all the operands for the given root node
		operandNodes := CreateNodePQ()
		operandSigns := map[*Node]bool{}
		// Collect all the operations along the traversal, and later on rebuild the
		// tree using these nodes
		operationNodes := []*Node{}

		var flatten func(n *Node, op NodeOp, sign bool) int
		var rebuild func(n *Node)

		defer func() {
			// Recursively collect all operands
			ranks[root] = flatten(root.Fanin(0), root.op, root.FaninSign(0)) +
				flatten(root.Fanin(1), root.op, root.FaninSign(1))
			// Build a balanced tree for this tree root
			rebuild(root)
		}()

		/*
		   Find all operands for a sub-tree starting with node n
		*/
		flatten = func(n *Node, op NodeOp, sign bool) int {
			if ranks[n] >= 0 {
				// This node is already processed, so it becomes an operand
				operandNodes.Push(NodePQEntry{n, ranks[n]})
				operandSigns[n] = sign
			} else if n.kind == NodeKind_Constant {
				// A constant has rank 0 and it's an operand
				ranks[n] = 0
				operandNodes.Push(NodePQEntry{n, ranks[n]})
				operandSigns[n] = sign
			} else if n.kind == NodeKind_Input || n.op != op {
				// Reach the boundary of the sub-tree, either input or a node with
				// different operation, and it's an operand
				ranks[n] = 1
				operandNodes.Push(NodePQEntry{n, ranks[n]})
				operandSigns[n] = sign
			} else if exist := candidateRoots.FindNode(n); exist {
				// If the node is also a candidate tree root, build it recursively and
				// it becomes an operand
				balance(n)
				operandNodes.Push(NodePQEntry{n, ranks[n]})
				operandSigns[n] = sign
			} else {
				// An internal node in a sub-tree, recursively find its operands
				ranks[n] = flatten(n.Fanin(0), n.op, n.FaninSign(0)) +
					flatten(n.Fanin(1), n.op, n.FaninSign(1))
				operationNodes = append(operationNodes, n)
			}

			return ranks[n]
		}

		/*
		   Build a balanced tree for a tree starting with the given root node
		*/
		rebuild = func(root *Node) {
			// Two operands mean there's only one operation, so no need to rebuild
			if operandNodes.Len() == 2 {
				return
			}

			// Disconnect the root from its fanins
			for root.NumFanins() > 0 {
				root.Fanin(0).RemoveFanout(root)
				root.RemoveFanin(root.Fanin(0))
			}
			// Disconnect operation nodes from their fanins and fanouts
			for _, node := range operationNodes {
				for node.NumFanins() > 0 {
					node.Fanin(0).RemoveFanout(node)
					node.RemoveFanin(node.Fanin(0))
				}
				for node.NumFanouts() > 0 {
					node.Fanout(0).RemoveFanin(node)
					node.RemoveFanout(node.Fanout(0))
				}
			}

			// At this point, we have a bunch of operands (in operandNodes) and a
			// bunch of operation nodes (in operationNodes and also root), now let's
			// build a balanced tree using these operation nodes for the operands

			for operandNodes.Len() > 0 {
				// Combine operands with the lowest ranks in the queue
				var nodeL *Node = operandNodes.PopMin()
				var nodeR *Node = operandNodes.PopMin()

				var nodeT *Node
				if operandNodes.Len() == 0 {
					// We've reached the root
					nodeT = root
				} else {
					// Pop one operation node
					nodeT = operationNodes[len(operationNodes)-1]
					operationNodes = operationNodes[:len(operationNodes)-1]
				}

				// Connect operands to operation node
				nodeT.Receive(nodeL)
				nodeT.Receive(nodeR)

				if operandSigns[nodeL] {
					nodeT.NegateFaninByIndex(0)
				}
				if operandSigns[nodeR] {
					nodeT.NegateFaninByIndex(1)
				}

				// Calculate operation node's rank
				ranks[nodeT] = ranks[nodeL] + ranks[nodeR]

				if operandNodes.Len() != 0 {
					// The operation node now becomes an operand for succeding operations
					operandNodes.Push(NodePQEntry{nodeT, ranks[nodeT]})
				}
			}
		} // func rebuild
	} // func balance
}
