package forge

import "fmt"

type Scheduler struct {
	graph *Graph

	compCost map[NodeOp]int
	commCost [][]int

	avgCompCost map[NodeOp]float64
	avgCommCost [][]float64

	processorInfo map[NodeOp][]int
	processorSlot [][]*Node
	processorBuf0 [][]int
	processorBuf1 [][]int
}

/*
ScheduleHEFT schedules operations in the graph using the Heterogeneous Earliest
Finish Time (HEFT) algorithm.

See "Performance-Effective and Low-Complexity Task Scheduling for Heterogeneous
Computing" by Haluk Topcuoglu et al. for details.
*/
func (s *Scheduler) ScheduleHEFT() {
	for _, node := range s.graph.allNodes {
		switch node.kind {
		case NodeKind_Input, NodeKind_Constant:
			node.actualStartTime = 0
			node.actualFinishTime = 0
		default:
			node.actualStartTime = 1000000
			node.actualFinishTime = 1000000
		}
	}

	// Step 1: Set the computation costs of tasks and communication costs of edges
	//         with mean values.
	s.avgCompCost = make(map[NodeOp]float64)
	s.avgCompCost[NodeOp_Add] = float64(s.compCost[NodeOp_Add]) / float64(len(s.processorInfo[NodeOp_Add]))
	s.avgCompCost[NodeOp_Mul] = float64(s.compCost[NodeOp_Mul]) / float64(len(s.processorInfo[NodeOp_Mul]))
	s.avgCompCost[NodeOp_Div] = float64(s.compCost[NodeOp_Div]) / float64(len(s.processorInfo[NodeOp_Div]))
	s.avgCompCost[NodeOp_Sin] = float64(s.compCost[NodeOp_Sin]) / float64(len(s.processorInfo[NodeOp_Sin]))
	s.avgCompCost[NodeOp_Cos] = s.avgCompCost[NodeOp_Sin]
	s.avgCompCost[NodeOp_Power] = s.avgCompCost[NodeOp_Mul]

	// Step 2: Compute rank_u for all tasks by traversing graph upward, starting
	//         from the exist tasks.
	maxLevel := s.graph.Levelize()

	maxURank := -100.0
	uRanks := make(map[*Node]float64)

	pq := CreateNodePQ()
	for _, node := range s.graph.operationNodes {
		pq.Push(NodePQEntry{node, maxLevel - node.level})
	}

	for pq.Len() > 0 {
		node := pq.Pop()

		uRank := -100.0
		for _, fo := range node.fanouts {
			foURank := 0.0
			if fo.kind != NodeKind_Output {
				foURank += uRanks[fo]
			}
			//foURank += 1 // comm

			if foURank > uRank {
				uRank = foURank
			}
		}
		uRank += s.avgCompCost[node.op]

		uRanks[node] = uRank

		if uRank > maxURank {
			maxURank = uRank
		}
	}

	// Step 3: Sort the tasks in a scheduling list by nonincreasing order of
	//         rank_u values.
	for node, uRank := range uRanks {
		pq.Push(NodePQEntry{node, maxURank - uRank})
	}

	// Step 4: While there are unscheduled tasks in the list do
	for pq.Len() > 0 {
		// Step 5: Select the first task from the list for scheduling.
		node := pq.Pop()

		minEST := 1000000
		minEFT := 1000000
		minEFTProcessor := -1

		// Step 6: For each processor in the processor-set do
		for _, processor := range s.processorInfo[node.op] {
			// Step 7: Compute EFT value using the insertion-based scheduling policy.
			EST, EFT := s.computeEFT(node, processor)

			if EFT < minEFT {
				minEST = EST
				minEFT = EFT
				minEFTProcessor = processor
			}
		}

		// Step 8: Assign task to the processor that minimizes EFT of task.
		s.processorSlot[minEFTProcessor][minEST] = node

		for t := node.Fanin(0).actualFinishTime; t < minEST; t++ {
			s.processorBuf0[minEFTProcessor][t]++
		}
		if node.NumFanins() > 1 {
			for t := node.Fanin(1).actualFinishTime; t < minEST; t++ {
				s.processorBuf1[minEFTProcessor][t]++
			}
		}

		node.actualStartTime = minEST
		node.actualFinishTime = minEFT
		node.processorAssigned = minEFTProcessor
	}

	//for p := 0; p < len(s.processorSlot); p++ {
	//	buf0Max := -100
	//	buf1Max := -100

	//	for _, v := range s.processorBuf0[p] {
	//		if v > buf0Max {
	//			buf0Max = v
	//		}
	//	}
	//	for _, v := range s.processorBuf1[p] {
	//		if v > buf1Max {
	//			buf1Max = v
	//		}
	//	}

	//	fmt.Println(p, buf0Max, buf1Max)
	//}
}

/*
computeEFT computes the earliest finish time for a node that can be scheduled
to a list of processors.
*/
func (s *Scheduler) computeEFT(node *Node, processor int) (int, int) {
	EST := 0

	for _, fi := range node.fanins {
		dataReadyTime := fi.actualFinishTime // + comm

		if dataReadyTime > EST {
			EST = dataReadyTime
		}
	}

	// Temporary
	if EST > len(s.processorSlot[processor]) {
		return EST, EST
	}

	for s.processorSlot[processor][EST] != nil {
		EST++
	}

	return EST, EST + s.compCost[node.op]
}

/*
Schedule performs heuristic scheduling.
*/
func (s *Scheduler) Schedule() {
	externalNodes := make(map[string]int)
	for key, _ := range s.graph.inputNodes {
		externalNodes[key] = len(externalNodes)
	}
	for key, _ := range s.graph.constantNodes {
		externalNodes[key] = len(externalNodes)
	}
	numExternalNodes := len(externalNodes) + 1

	roots := CreateNodePQ()

	rootInputs := make(map[string][]bool)

	// Find all tree roots
	for _, node := range s.graph.operationNodes {
		// A tree root has either multiple fanouts, or a single fanout to an output
		if (node.NumFanouts() > 1) ||
			(node.NumFanouts() == 1 && node.Fanout(0).kind == NodeKind_Output) {
			maxFaninLevel := 0
			sumInputFanouts := 0

			traversedExternalNodes := make(map[string]bool)

			rootInputs[node.name] = make([]bool, numExternalNodes)

			var traverse func(*Node)
			traverse = func(n *Node) { // must use a name different from "node"
				for _, fanin := range n.fanins {
					if fanin.kind == NodeKind_Input || fanin.kind == NodeKind_Constant {
						// External input or constant
						if _, exist := traversedExternalNodes[fanin.name]; !exist {
							//fmt.Printf(" %s", fanin.name)

							traversedExternalNodes[fanin.name] = true
							sumInputFanouts += fanin.NumFanouts()

							rootInputs[node.name][externalNodes[fanin.name]] = true
						}
					} else if fanin.NumFanouts() > 1 {
						// Another tree root
						//fmt.Printf(" %s", fanin.name)

						if fanin.level > maxFaninLevel {
							maxFaninLevel = fanin.level
						}
					} else {
						traverse(fanin)
					}
				}
			}

			//fmt.Printf("%s @%d   ", node.name, node.level)
			traverse(node)
			//fmt.Printf(" | %d %d\n", maxFaninLevel, sumInputFanouts)

			roots.Push(NodePQEntry{node, -(1000*(100-maxFaninLevel) + sumInputFanouts)})
		}
	}

	//for i := 0; i < roots.Len(); i++ {
	//  nodeI := roots.GetNodeByIndex(i)

	//  maxNumCommonInput := -1
	//  similarRoot := ""

	//  for j := 0; j < roots.Len(); j++ {
	//    if (i != j) {
	//      nodeJ := roots.GetNodeByIndex(j)

	//      numCommonInput := 0

	//      for idx := 0; idx < numExternalNodes; idx++ {
	//        if rootInputs[nodeI.name][idx] && rootInputs[nodeJ.name][idx] {
	//          numCommonInput++
	//        }
	//      }

	//      if numCommonInput > maxNumCommonInput {
	//        maxNumCommonInput = numCommonInput
	//        similarRoot = nodeJ.name
	//      }
	//    }
	//  }

	//  fmt.Printf("%s ~= %s (%d)\n", nodeI.name, similarRoot, maxNumCommonInput)
	//}
	//fmt.Printf("\n")

	for roots.Len() > 0 {
		var list []*Node

		entry := roots.PopEntry()
		list = append(list, entry.Payload)

		priority := entry.Priority.(int)
		//fmt.Printf("%s ", entry.Payload.name)

		for roots.Len() > 0 {
			entry := roots.PopEntry()

			if priority != entry.Priority.(int) {
				roots.Push(entry)
				break
			}

			//fmt.Printf("%s ", entry.Payload.name)
			list = append(list, entry.Payload)
		}

		fmt.Printf("List length = %d\n", len(list))

		for len(list) > 0 {
			nodeI := list[0]
			fmt.Printf("%s\n", nodeI.name)

			maxNumCommonInput := -1
			//similarRoot := ""
			similarRootIdx := -1

			for j := 1; j < len(list); j++ {
				nodeJ := list[j]

				numCommonInput := 0

				for idx := 0; idx < numExternalNodes; idx++ {
					if rootInputs[nodeI.name][idx] && rootInputs[nodeJ.name][idx] {
						numCommonInput++
					}
				}

				if numCommonInput > maxNumCommonInput {
					maxNumCommonInput = numCommonInput
					//similarRoot = nodeJ.name
					similarRootIdx = j
				}
			}

			if len(list) == 1 {
				break
			} else {
				list[0] = list[similarRootIdx]
				list = append(list[:similarRootIdx], list[similarRootIdx+1:]...)
			}

			//fmt.Printf("%s ~= %s (%d)\n", nodeI.name, similarRoot, maxNumCommonInput)
		}
		//fmt.Printf("\n")

		//node, priority := entry.Payload, entry.Priority.(int)

		//fmt.Printf("%s %d, ", node.name, priority)
		//for i := 0; i < numExternalNodes; i++ {
		//  if rootInputs[node.name][i] {
		//    fmt.Printf("1")
		//  } else {
		//    fmt.Printf("0")
		//  }
		//}
		//fmt.Printf("\n")
	}
}
