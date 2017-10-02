package forge

import "fmt"

type PEKind int

const (
	PEKind_Add PEKind = iota
	PEKind_Mul
	PEKind_Div
)

// -----------------------------------------------------------------------------

type ProcessingElement struct {
	kind PEKind

	dataSlots0 [][]*Node
	dataSlots1 [][]*Node

	executionSlots []*Node
}

// -----------------------------------------------------------------------------

type ProcessingGroup struct {
	inputSlots [][]*Node

	processingElements []*ProcessingElement
}

func (pg *ProcessingGroup) AddInputSlots(numSlots int) {
	pg.inputSlots = make([][]*Node, numSlots)

	for i := 0; i < numSlots; i++ {
		pg.inputSlots[i] = make([]*Node, 32767)
	}
}

func (pg *ProcessingGroup) AddProcessingElement(pe *ProcessingElement) {
	pg.processingElements = append(pg.processingElements, pe)
}

//func (pg *ProcessingGroup) CanAllocateInput(n *Node, time int) bool {
//	for line := 0; line < len(pg.inputSlots); line++ {
//		if pg.inputSlots[line][time] == nil {
//			return true
//		}
//	}
//
//	return false
//}

func (pg *ProcessingGroup) GetEarliestInputSlot(startLine, startTime int) (int, int) {
	if startLine == -1 && startTime == -1 {
		startLine = 0
		startTime = 0
	} else {
		if startLine == len(pg.inputSlots)-1 {
			startLine = 0
			startTime++
		} else {
			startLine++
		}
	}

	for time := startTime; time < 32767; time++ {
		for line := startLine; line < len(pg.inputSlots); line++ {
			if pg.inputSlots[line][time] == nil {
				return line, time
			}
		}
	}

	return -1, -1
}

func (pg *ProcessingGroup) AllocateInput(n *Node, time int) {
	for line := 0; line < len(pg.inputSlots); line++ {
		if pg.inputSlots[line][time] == nil {
			pg.inputSlots[line][time] = n
			break
		}
	}
}

// -----------------------------------------------------------------------------

type Processor struct {
	processingGroups []*ProcessingGroup
}

func (p *Processor) AddProcessingGroup(pg *ProcessingGroup) {
	p.processingGroups = append(p.processingGroups, pg)
}

// -----------------------------------------------------------------------------

func (s *Scheduler) ConfigureHW() {
	s.processor = &Processor{}

	for g := 0; g < 2; g++ {
		pg := &ProcessingGroup{}

		pg.AddInputSlots(2)

		for e := 0; e < 5; e++ {
			pe := &ProcessingElement{}

			if e < 2 {
				pe.kind = PEKind_Add
			} else if e < 4 {
				pe.kind = PEKind_Mul
			} else {
				pe.kind = PEKind_Div
			}

			pe.dataSlots0 = make([][]*Node, 128)
			pe.dataSlots1 = make([][]*Node, 128)
			for i := 0; i < 128; i++ {
				pe.dataSlots0[i] = make([]*Node, 32767)
				pe.dataSlots1[i] = make([]*Node, 32767)
			}

			pe.executionSlots = make([]*Node, 32767)

			pg.AddProcessingElement(pe)
		}

		s.processor.AddProcessingGroup(pg)
	}

	fmt.Println("")

	//s.compCost = make(map[NodeOp]int)
	//s.compCost[NodeOp_Add] = 1
	//s.compCost[NodeOp_Mul] = 1
	//s.compCost[NodeOp_Div] = 3
	//s.compCost[NodeOp_Sin] = 5
	//s.compCost[NodeOp_Cos] = s.compCost[NodeOp_Sin]
	//s.compCost[NodeOp_Power] = s.compCost[NodeOp_Mul]

	//s.processorInfo = make(map[NodeOp][]int)
	//s.processorInfo[NodeOp_Add] = []int{0}
	//s.processorInfo[NodeOp_Mul] = []int{1}
	//s.processorInfo[NodeOp_Div] = []int{2}
	//s.processorInfo[NodeOp_Sin] = []int{3}
	//s.processorInfo[NodeOp_Cos] = s.processorInfo[NodeOp_Sin]
	//s.processorInfo[NodeOp_Power] = s.processorInfo[NodeOp_Mul]

	//numProcessors := 4

	//s.processorSlot = make([][]*Node, numProcessors)
	//s.processorBuf0 = make([][]int, numProcessors)
	//s.processorBuf1 = make([][]int, numProcessors)
	//for p := 0; p < numProcessors; p++ {
	//	s.processorSlot[p] = make([]*Node, 5*s.graph.NumOperationNodes())
	//	s.processorBuf0[p] = make([]int, 5*s.graph.NumOperationNodes())
	//	s.processorBuf1[p] = make([]int, 5*s.graph.NumOperationNodes())
	//}
}
