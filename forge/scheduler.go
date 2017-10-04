package forge

import (
	"fmt"
	"strconv"
)

type Scheduler struct {
	graph       *Graph
	mergedGraph *Graph

	processor *Processor
}

/*
ScheduleHeuristic schedules nodes onto the hardware heuristically.
*/
func (s *Scheduler) ScheduleHeuristic() {
	// This maps the name of an external node (input or constant) to an identification
	// number for quick search and comparison.
	extNodeIds := make(map[string]int)
	for key, _ := range s.graph.inputNodes {
		extNodeIds[key] = len(extNodeIds)
	}
	for key, _ := range s.graph.constantNodes {
		extNodeIds[key] = len(extNodeIds)
	}
	numExtNodes := len(extNodeIds)

	// Create a priority queue to store roots in the graph.
	roots := CreateNodePQ()

	// This maps the name of a root node to a list of boolean values. Each
	// boolean value represents if the corresponding external node is used by
	// the root node or not.
	rootExtNodes := make(map[string][]bool)

	// Stage 1: Partition the graph into sub-graphs and assign priority to these
	//          sub-graphs based on their level and external node usage.
	// ---------------------------------------------------------------------------

	//fmt.Println(" Partitioning the graph ...")

	// Find root of the sub-graphs.
	for _, node := range s.graph.operationNodes {
		// A root has either multiple fanouts, or a single fanout to an output.
		if (node.NumFanouts() > 1) ||
			(node.NumFanouts() == 1 && node.Fanout(0).kind == NodeKind_Output) {
			// Now we have a root node, we need to traverse it and calculate some numbers.

			// Allocate a list to keep track of external nodes this root uses.
			rootExtNodes[node.name] = make([]bool, numExtNodes)

			// Maximum level of input to this root.
			maxInputLevel := 0
			// Total number of fanouts of the external nodes that are used by this root.
			sumExtNodeFanouts := 0

			// Keep track of traversed external nodes to avoid counting the same external
			// node multiple times, otherwise sub-graphs using multi-fanout external nodes
			// multiple times will be given extra high priority.
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

							rootExtNodes[node.name][extNodeIds[fanin.name]] = true
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
			// Now do the actual traversal.
			traverse(node)

			// Calculate the priority of this sub-graph. Smaller maximum input level is
			// given higher priority because it's closer to the input so should be
			// computed earlier. If two sub-graphs have the same maximum input level,
			// then the sum of their external node's fanout breaks the tie. The overall
			// priority is negated because the priority queue is ordered small to large.
			priority := -(1000*(100-maxInputLevel) + sumExtNodeFanouts)

			roots.Push(NodePQEntry{node, priority})
		}
	}

	// At this point, the priority queue roots contains root of sub-graphs sorted
	// by their priority.

	// Stage 2: Schedule the sub-graphs.
	// ---------------------------------------------------------------------------

	//fmt.Println(" Scheduling the graph ...")

	inputMap := make(map[string]int)

	costTblScheduleTime := make([][]int, len(s.processor.processGroups))
	for pgId, pg := range s.processor.processGroups {
		costTblScheduleTime[pgId] = make([]int, len(pg.processElements))
	}

	finalFinishTime := 0

	// While there's still unscheduled roots.
	for roots.Len() > 0 {
		// Get the highest priority root.
		entry := roots.PopEntry()
		priority := entry.Priority.(int)

		// A list of roots having the same priority. These roots will be processed
		// in the order of their similarity.
		var samePriorityRoots []*Node
		samePriorityRoots = append(samePriorityRoots, entry.Payload)
		// Now find all roots having the same priority.
		for roots.Len() > 0 {
			entry := roots.PopEntry()

			// If the next root has a different priority, push it back and we're done.
			if priority != entry.Priority.(int) {
				roots.Push(entry)
				break
			}

			samePriorityRoots = append(samePriorityRoots, entry.Payload)
		}

		//fmt.Printf("List length = %d\n", len(samePriorityRoots))

		// Now we have a list of one or more roots, having the same priority.
		// We start from the first one, then the one having the most common external
		// nodes with it, and so on.
		for len(samePriorityRoots) > 0 {
			root := samePriorityRoots[0]
			//fmt.Printf("%s\n", root.name)

			var scheduleSubGraph func(*Node)
			scheduleSubGraph = func(n *Node) {
				// DFS
				for _, fanin := range n.fanins {
					if fanin.NumFanouts() == 1 &&
						fanin.kind != NodeKind_Input && fanin.kind != NodeKind_Constant {
						scheduleSubGraph(fanin)
					}
				}

				// Initialize cost table for schedule time.
				for pgId, pg := range s.processor.processGroups {
					for peId, _ := range pg.processElements {
						costTblScheduleTime[pgId][peId] = MaxInt
					}
				}

				earliestArrivalTime := MaxInt
				latestArrivalTime := 0

				for pgId := 0; pgId < len(s.processor.processGroups); pgId++ {
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
				} // end of searching process group

				// Find the best process group/element to schedule this node based on the
				// cost table.
				bestPGId := -1
				bestPEId := -1
				scheduleTime := MaxInt
				for pgId, pg := range s.processor.processGroups {
					for peId, _ := range pg.processElements {
						if costTblScheduleTime[pgId][peId] < scheduleTime {
							bestPGId = pgId
							bestPEId = peId
							scheduleTime = costTblScheduleTime[pgId][peId]
						}
					}
				}
				bestPG := s.processor.processGroups[bestPGId]
				bestPE := bestPG.processElements[bestPEId]

				// Schedule external nodes that have not been scheduled yet.
				inputLine := -1
				inputTime := -1
				for _, fanin := range n.fanins {
					if fanin.kind == NodeKind_Input || fanin.kind == NodeKind_Constant {
						key := fanin.name + "@" + strconv.FormatInt(int64(bestPGId), 10)

						if _, exist := inputMap[key]; !exist {
							inputLine, inputTime = bestPG.GetEarliestInputSlot(inputLine, inputTime)

							inputMap[key] = inputTime
							bestPG.AllocateInput(n, inputTime)
						}
					}
				}

				// Schedule this node.
				if bestPE.executionSlots[scheduleTime] != nil {
					fmt.Printf("ERROR: execution slot is occupied!")
				}

				bestPE.executionSlots[scheduleTime] = n
				n.isScheduled = true
				n.pgScheduled = bestPGId
				n.peScheduled = bestPEId
				n.startTime = scheduleTime
				switch n.op {
				case NodeOp_Add:
					n.finishTime = scheduleTime + 1
				case NodeOp_Mul, NodeOp_Power:
					n.finishTime = scheduleTime + 2
				case NodeOp_Div:
					n.finishTime = scheduleTime + 3
				case NodeOp_Sin, NodeOp_Cos:
					n.finishTime = scheduleTime + 3
				default:
					fmt.Printf("ERROR: %s has unsupported operation %d\n", n.name, n.op)
				}

				if n.finishTime > finalFinishTime {
					finalFinishTime = n.finishTime
				}

				//fmt.Printf(" schedule %s to G%d E%d @%d\n", n.name, bestPGId, bestPEId, scheduleTime)
			}
			scheduleSubGraph(root)

			// If this is the last one then we're done scheduling roots with the same
			// priority, otherwise look for another same priority root sharing the most
			// common external nodes with this root.
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

			// Relocate the most similar root to the front so it will be picked in the
			// next iteration.
			samePriorityRoots[0] = samePriorityRoots[similarRootIdx]
			samePriorityRoots = append(
				samePriorityRoots[:similarRootIdx], samePriorityRoots[similarRootIdx+1:]...)
		}
	}

	fmt.Printf("  %d nodes in %d cycles, speedup = %.2f\n",
		len(s.graph.operationNodes), finalFinishTime,
		float32(len(s.graph.operationNodes))/float32(finalFinishTime))

	if s.mergedGraph != nil {
		fmt.Printf(" (%d nodes in %d cycles, speedup = %.2f)\n",
			len(s.mergedGraph.operationNodes), finalFinishTime,
			float32(len(s.mergedGraph.operationNodes))/float32(finalFinishTime))
	}

	fmt.Printf("\n")

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
