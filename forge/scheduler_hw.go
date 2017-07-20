package forge

//import "fmt"

func (s *Scheduler) ConfigureHW() {
	s.compCost = make(map[NodeOp]int)
	s.compCost[NodeOp_Add] = 1
	s.compCost[NodeOp_Mul] = 1
	s.compCost[NodeOp_Div] = 3
	s.compCost[NodeOp_Sin] = 5
	s.compCost[NodeOp_Cos] = s.compCost[NodeOp_Sin]
	s.compCost[NodeOp_Power] = s.compCost[NodeOp_Mul]

	s.processorInfo = make(map[NodeOp][]int)
	s.processorInfo[NodeOp_Add] = []int{0}
	s.processorInfo[NodeOp_Mul] = []int{1}
	s.processorInfo[NodeOp_Div] = []int{2}
	s.processorInfo[NodeOp_Sin] = []int{3}
	s.processorInfo[NodeOp_Cos] = s.processorInfo[NodeOp_Sin]
	s.processorInfo[NodeOp_Power] = s.processorInfo[NodeOp_Mul]

	numProcessors := 4

	s.processorSlot = make([][]*Node, numProcessors)
	s.processorBuf0 = make([][]int, numProcessors)
	s.processorBuf1 = make([][]int, numProcessors)
	for p := 0; p < numProcessors; p++ {
		s.processorSlot[p] = make([]*Node, 5*s.graph.NumOperationNodes())
		s.processorBuf0[p] = make([]int, 5*s.graph.NumOperationNodes())
		s.processorBuf1[p] = make([]int, 5*s.graph.NumOperationNodes())
	}
}
