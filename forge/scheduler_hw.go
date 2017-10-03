package forge

//import "fmt"

// -----------------------------------------------------------------------------

func (s *Scheduler) ConfigureHW() {
	s.processor = &Processor{}

	// Basic operation process groups.
	for g := 0; g < 2; g++ {
		pg := &ProcessGroup{}

		pg.AddInputSlots(2)

		for e := 0; e < 5; e++ {
			pe := CreateProcessElement()

			if e < 2 {
				pe.kind = ProcessElementKind_Add
			} else if e < 4 {
				pe.kind = ProcessElementKind_Mul
			} else {
				pe.kind = ProcessElementKind_Div
			}

			pg.AddProcessElement(pe)
		}

		s.processor.AddProcessGroup(pg)
	}

	// Sinusoid operation process groups.
	for g := 0; g < 1; g++ {
		pg := &ProcessGroup{}

		pg.AddInputSlots(2)

		for e := 0; e < 2; e++ {
			pe := CreateProcessElement()

			pe.kind = ProcessElementKind_Cordic

			pg.AddProcessElement(pe)
		}

		s.processor.AddProcessGroup(pg)
	}
}
