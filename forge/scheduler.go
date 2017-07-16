package forge

//import "fmt"

type Scheduler struct {
	graph *Graph

	compCost map[NodeOp]int
	commCost [][]int
}

/*
ScheduleHEFT schedules operations in the graph using the Heterogeneous Earliest
Finish Time (HEFT) algorithm.

See "Performance-Effective and Low-Complexity Task Scheduling for Heterogeneous
Computing" by Haluk Topcuoglu et al. for details.
*/
func (s *Scheduler) ScheduleHEFT() {
	// Step 1: Set the computation costs of tasks and communication costs of edges
	//         with mean values.

	// Step 2: Compute rank_u for all tasks by traversing graph upward, starting
	//         from the exist tasks.
	maxLevel := s.graph.Levelize()

	maxURank := -100
	uRanks := make(map[*Node]int)

	pq := CreateNodePQ()
	for _, node := range s.graph.operationNodes {
		pq.Push(NodePQEntry{node, maxLevel - node.level})
	}

	for pq.Len() > 0 {
		node := pq.Pop()

		uRank := -100
		for _, fo := range node.fanouts {
			foURank := 0
			if fo.kind != NodeKind_Output {
				foURank += uRanks[fo]
			}
			//foURank += 1 // comm

			if foURank > uRank {
				uRank = foURank
			}
		}
		uRank += 1 // comp

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

		// Step 6: For each processor in the processor-set do

		// Step 7: Compute EFT value using the insertion-based scheduling policy.

		// Step 8: Assign task to the processor that minimizes EFT of task.
		node.name = node.name
	}
}
