package forge

import "fmt"

type ProcessElementKind int

const tmpMaxCycle = 32767

const (
	ProcessElementKind_Add ProcessElementKind = iota
	ProcessElementKind_Mul
	ProcessElementKind_Div
	ProcessElementKind_Cordic
)

var compatibleMap = make(map[NodeOp][][]int)

// -----------------------------------------------------------------------------

type ProcessElement struct {
	kind ProcessElementKind

	buffer0 [][]*Node
	buffer1 [][]*Node

	executionSlots []*Node
}

type ProcessGroup struct {
	inputSlots [][]*Node

	processElements []*ProcessElement
}

type Processor struct {
	processGroups []*ProcessGroup
}

// -----------------------------------------------------------------------------

func CreateProcessElement() *ProcessElement {
	pe := &ProcessElement{}

	//pe.buffer0 = make([][]*Node, 128)
	//pe.buffer1 = make([][]*Node, 128)
	//for i := 0; i < 128; i++ {
	//	pe.buffer0[i] = make([]*Node, tmpMaxCycle)
	//	pe.buffer1[i] = make([]*Node, tmpMaxCycle)
	//}

	pe.executionSlots = make([]*Node, tmpMaxCycle)

	return pe
}

// -----------------------------------------------------------------------------

func (pg *ProcessGroup) AddInputSlots(numSlots int) {
	pg.inputSlots = make([][]*Node, numSlots)

	for i := 0; i < numSlots; i++ {
		pg.inputSlots[i] = make([]*Node, tmpMaxCycle)
	}
}

func (pg *ProcessGroup) AddProcessElement(pe *ProcessElement) {
	pg.processElements = append(pg.processElements, pe)
}

func (pg *ProcessGroup) GetEarliestInputSlot(startLine, startTime int) (int, int) {
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

	for time := startTime; time < tmpMaxCycle; time++ {
		for line := startLine; line < len(pg.inputSlots); line++ {
			if pg.inputSlots[line][time] == nil {
				return line, time
			}
		}
	}

	return -1, -1
}

func (pg *ProcessGroup) AllocateInput(n *Node, time int) {
	for line := 0; line < len(pg.inputSlots); line++ {
		if pg.inputSlots[line][time] == nil {
			pg.inputSlots[line][time] = n
			break
		}
	}
}

// -----------------------------------------------------------------------------

func CreateProcessor() *Processor {
	processor := &Processor{}

	// Basic operation process groups.
	for g := 0; g < 16; g++ {
		pg := &ProcessGroup{}

		pg.AddInputSlots(16)

		for e := 0; e < 18; e++ {
			pe := CreateProcessElement()

			if e < 8 {
				pe.kind = ProcessElementKind_Add
			} else if e < 16 {
				pe.kind = ProcessElementKind_Mul
			} else {
				pe.kind = ProcessElementKind_Div
			}

			pg.AddProcessElement(pe)
		}

		processor.AddProcessGroup(pg)
	}

	// Sinusoid operation process groups.
	for g := 0; g < 2; g++ {
		pg := &ProcessGroup{}

		pg.AddInputSlots(4)

		for e := 0; e < 4; e++ {
			pe := CreateProcessElement()

			pe.kind = ProcessElementKind_Cordic

			pg.AddProcessElement(pe)
		}

		processor.AddProcessGroup(pg)
	}

	return processor
}

func (p *Processor) AddProcessGroup(pg *ProcessGroup) {
	pgId := len(p.processGroups)
	p.processGroups = append(p.processGroups, pg)

	for _, op := range NodeOpLUT {
		compatibleMap[op] = append(compatibleMap[op], make([]int, 0))
	}

	for peId, pe := range pg.processElements {
		switch pe.kind {
		case ProcessElementKind_Add:
			{
				compatibleMap[NodeOp_Add][pgId] = append(compatibleMap[NodeOp_Add][pgId], peId)
			}
		case ProcessElementKind_Mul:
			{
				compatibleMap[NodeOp_Mul][pgId] = append(compatibleMap[NodeOp_Mul][pgId], peId)
				compatibleMap[NodeOp_Power][pgId] = append(compatibleMap[NodeOp_Power][pgId], peId)
			}
		case ProcessElementKind_Div:
			{
				compatibleMap[NodeOp_Div][pgId] = append(compatibleMap[NodeOp_Div][pgId], peId)
			}
		case ProcessElementKind_Cordic:
			{
				compatibleMap[NodeOp_Sin][pgId] = append(compatibleMap[NodeOp_Sin][pgId], peId)
				compatibleMap[NodeOp_Cos][pgId] = append(compatibleMap[NodeOp_Cos][pgId], peId)
			}
		default:
			fmt.Printf("ERROR")
		}
	}
}
