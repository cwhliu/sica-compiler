package forge

//import "fmt"

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
