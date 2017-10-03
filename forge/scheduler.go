package forge

import (
	"fmt"
	"strconv"
)

type Scheduler struct {
	graph *Graph

	processor *Processor

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

	// Step 3: Sort the tasks in a scheduling samePriorityRoots by nonincreasing order of
	//         rank_u values.
	for node, uRank := range uRanks {
		pq.Push(NodePQEntry{node, maxURank - uRank})
	}

	// Step 4: While there are unscheduled tasks in the samePriorityRoots do
	for pq.Len() > 0 {
		// Step 5: Select the first task from the samePriorityRoots for scheduling.
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
to a samePriorityRoots of processors.
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
	// This maps the name of an external node (input or constant) to a number
	// for quick search and comparison.
	extNodeIDs := make(map[string]int)
	for key, _ := range s.graph.inputNodes {
		extNodeIDs[key] = len(extNodeIDs)
	}
	for key, _ := range s.graph.constantNodes {
		extNodeIDs[key] = len(extNodeIDs)
	}
	numExtNodes := len(extNodeIDs) + 1

	roots := CreateNodePQ()

	// This maps the name of a root node to a samePriorityRoots of boolean values. Each
	// boolean value represents if the external node is used by the root node.
	rootExtNodes := make(map[string][]bool)

	fmt.Println("Partitioning the graph ...")

	// Find all roots.
	for _, node := range s.graph.operationNodes {
		// A root has either multiple fanouts, or a single fanout to an output.
		if (node.NumFanouts() > 1) ||
			(node.NumFanouts() == 1 && node.Fanout(0).kind == NodeKind_Output) {
			// Now we have a root node, we need to traverse it and calculate some numbers.

			// Allocate a samePriorityRoots for noting external nodes this root uses.
			rootExtNodes[node.name] = make([]bool, numExtNodes)

			// Maximum level of input to this root.
			maxInputLevel := 0
			// Total number of fanouts of the external nodes that are used by this root.
			sumExtNodeFanouts := 0

			// Keep track of traversed external nodes to avoid duplication.
			traversedExtNodes := make(map[string]bool)

			// This function recursively traverses a given node back to its inputs.
			var traverse func(*Node)
			traverse = func(n *Node) {
				for _, fanin := range n.fanins {
					if fanin.kind == NodeKind_Input || fanin.kind == NodeKind_Constant {
						if _, exist := traversedExtNodes[fanin.name]; !exist {
							// fanin is a not yet traversed external nodes.

							traversedExtNodes[fanin.name] = true
							sumExtNodeFanouts += fanin.NumFanouts()

							rootExtNodes[node.name][extNodeIDs[fanin.name]] = true
						}
					} else if fanin.NumFanouts() > 1 {
						// fanin is another root.

						if fanin.level > maxInputLevel {
							maxInputLevel = fanin.level
						}
					} else {
						traverse(fanin)
					}
				}
			}
			traverse(node)

			priority := -(1000*(100-maxInputLevel) + sumExtNodeFanouts)

			roots.Push(NodePQEntry{node, priority})
		}
	}

	fmt.Println("Scheduling the graph ...")

	inputMap := make(map[string]int)

	// roots now contains root nodes sorted by their priority, let's process them.
	for roots.Len() > 0 {
		// Get the first priority root.
		entry := roots.PopEntry()
		priority := entry.Priority.(int)

		// A list of roots having the same priority. These roots will be processed
		// in the order of similarity.
		var samePriorityRoots []*Node
		samePriorityRoots = append(samePriorityRoots, entry.Payload)

		// Now find all roots having the same priority.
		for roots.Len() > 0 {
			entry := roots.PopEntry()

			// If the next root has a different priority, push it back and we're done finding.
			if priority != entry.Priority.(int) {
				roots.Push(entry)
				break
			}

			samePriorityRoots = append(samePriorityRoots, entry.Payload)
		}

		fmt.Printf("List length = %d\n", len(samePriorityRoots))

		costTblScheduleTime := make([][]int, len(s.processor.processGroups))
		for pgId, pg := range s.processor.processGroups {
			costTblScheduleTime[pgId] = make([]int, len(pg.processElements))
		}

		// Now we have a list of one or more roots, having the same priority.
		// We start from the first one, then the one having the most common external
		// nodes with it, and so on.
		for len(samePriorityRoots) > 0 {
			root := samePriorityRoots[0]

			fmt.Printf("%s\n", root.name)
			var traverse func(*Node)
			traverse = func(n *Node) {
				// DFS
				for _, fanin := range n.fanins {
					if fanin.NumFanouts() == 1 &&
						fanin.kind != NodeKind_Input && fanin.kind != NodeKind_Constant {
						traverse(fanin)
					}
				}

				// Create cost table for schedule time.
				for pgId, pg := range s.processor.processGroups {
					for peId, _ := range pg.processElements {
						costTblScheduleTime[pgId][peId] = 32767
					}
				}

				earliestArrivalTime := 32767
				latestArrivalTime := 0

				pgSearchStart := 0
				pgSearchStop := len(s.processor.processGroups)
				for pgId := pgSearchStart; pgId < pgSearchStop; pgId++ {
					// Skip checking the process group if this node can not be processed.
					if len(compatibleMap[n.op][pgId]) == 0 {
						continue
					}

					pg := s.processor.processGroups[pgId]

					inputLine := -1
					inputTime := -1

					// Compute the earliest and latest arrival time if the node is scheduled
					// at this process group.
					for _, fanin := range n.fanins {
						if fanin.kind == NodeKind_Input || fanin.kind == NodeKind_Constant {
							key := fanin.name + "@" + strconv.FormatInt(int64(pgId), 10)

							if _, exist := inputMap[key]; !exist {
								inputLine, inputTime = pg.GetEarliestInputSlot(inputLine, inputTime)

								if inputTime < earliestArrivalTime {
									earliestArrivalTime = inputTime
								}
								if inputTime > latestArrivalTime {
									latestArrivalTime = inputTime
								}
							} else {
								if inputMap[key] < earliestArrivalTime {
									earliestArrivalTime = inputMap[key]
								}
								if inputMap[key] > latestArrivalTime {
									latestArrivalTime = inputMap[key]
								}
							}
						} else {
							arrivalTime := fanin.finishTime
							if fanin.pgScheduled != pgId {
								arrivalTime++
							}

							if arrivalTime < earliestArrivalTime {
								earliestArrivalTime = arrivalTime
							}
							if arrivalTime > latestArrivalTime {
								latestArrivalTime = arrivalTime
							}
						}
					}

					for _, peId := range compatibleMap[n.op][pgId] {
						pe := pg.processElements[peId]

						for time := latestArrivalTime; ; time++ {
							if pe.executionSlots[time] == nil {
								costTblScheduleTime[pgId][peId] = time
								break
							}
						}
					}
				} // end of searching PG

				bestPG := -1
				bestPE := -1
				scheduleTime := 32767
				for pgId, pg := range s.processor.processGroups {
					for peId, _ := range pg.processElements {
						if costTblScheduleTime[pgId][peId] < scheduleTime {
							bestPG = pgId
							bestPE = peId
							scheduleTime = costTblScheduleTime[pgId][peId]
						}
					}
				}

				pg := s.processor.processGroups[bestPG]
				pe := pg.processElements[bestPE]

				// Schedule external node fanin.
				inputLine := -1
				inputTime := -1
				for _, fanin := range n.fanins {
					if fanin.kind == NodeKind_Input || fanin.kind == NodeKind_Constant {
						key := fanin.name + "@" + strconv.FormatInt(int64(bestPG), 10)

						if _, exist := inputMap[key]; !exist {
							inputLine, inputTime = pg.GetEarliestInputSlot(inputLine, inputTime)

							inputMap[key] = inputTime
							pg.AllocateInput(n, inputTime)
						}
					}
				}

				// Schedule this node.
				if pe.executionSlots[scheduleTime] != nil {
					fmt.Printf("ERROR: execution slot is occupied!")
				}

				pe.executionSlots[scheduleTime] = n
				n.isScheduled = true
				n.pgScheduled = bestPG
				n.peScheduled = bestPE
				n.startTime = scheduleTime
				switch n.op {
				case NodeOp_Add:
					{
						n.finishTime = scheduleTime + 1
					}
				case NodeOp_Mul, NodeOp_Power:
					{
						n.finishTime = scheduleTime + 2
					}
				case NodeOp_Div:
					{
						n.finishTime = scheduleTime + 3
					}
				case NodeOp_Sin, NodeOp_Cos:
					{
						n.finishTime = scheduleTime + 3
					}
				default:
					fmt.Printf("ERROR: %s has unsupported operation %d\n", n.name, n.op)
				}

				fmt.Printf(" schedule %s to G%d E%d @%d\n", n.name, bestPG, bestPE, scheduleTime)
			}
			traverse(root)

			// We're done if there's only one root left.
			if len(samePriorityRoots) == 1 {
				break
			}

			// Maximum number of common external nodes.
			maxNumCommon := -1
			// Which other root is the most similar to this one.
			similarRootIdx := -1

			for i := 1; i < len(samePriorityRoots); i++ {
				node := samePriorityRoots[i]

				// Count how many common external nodes these two roots have.
				numCommonExtNodes := 0
				for idx := 0; idx < numExtNodes; idx++ {
					if rootExtNodes[root.name][idx] && rootExtNodes[node.name][idx] {
						numCommonExtNodes++
					}
				}

				if numCommonExtNodes > maxNumCommon {
					maxNumCommon = numCommonExtNodes
					similarRootIdx = i
				}
			}

			// Relocate the most similar root to the front.
			samePriorityRoots[0] = samePriorityRoots[similarRootIdx]
			samePriorityRoots = append(
				samePriorityRoots[:similarRootIdx], samePriorityRoots[similarRootIdx+1:]...)
		}
	}

	//for key, val := range inputMap {
	//	fmt.Printf("%s = %d\n", key, val)
	//}

	//for op, list := range compatibleMap {
	//  fmt.Printf("Node OP = %s\n", NodeOpStringLUT[op])
	//  for pgId, _ := range list {
	//    fmt.Printf("%d:", pgId)
	//    for peId, _ := range list[pgId] {
	//      fmt.Printf(" %d", list[pgId][peId])
	//    }
	//    fmt.Printf("\n")
	//  }
	//}
}
