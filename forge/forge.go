package forge

import "fmt"

/*
Forge is the main compiler instance.
*/
type Forge struct {
	parser    Parser
	scheduler Scheduler
}

/*
BuildGraph invokes the parser to parse the C++ source file and build a graph for it.
*/
func (f *Forge) BuildGraph(filename string) error {
	if g, err := f.parser.Parse(filename); err != nil {
		return err
	} else {
		g.EvaluateGolden(1)

		g.SimplifyArithmetic()
		g.EliminateDuplicatedOperation()
		g.MaximizeParallelism()
		g.DeleteUnusedNodes()

		g.EvaluateCompare()

		fmt.Printf(" %d operation nodes, %d levels\n", g.NumOperationNodes(), g.Levelize())

		//fanoutCounter := [1000]int{}
		//for _, node := range g.operationNodes {
		//  fanoutCounter[node.NumFanouts()]++
		//}
		//for i := 1; i < 1000; i++ {
		//  if (fanoutCounter[i] > 0) {
		//    fmt.Printf(" %d=%d", i, fanoutCounter[i])
		//  }
		//}
		//fmt.Printf("\n")

		// Pass the graph to the scheduler
		f.scheduler.graph = g

		return nil
	}
}

/*
ScheduleGraph schedules the operations in the graph onto the hardware accelerator.
*/
func (f *Forge) ScheduleGraph() {
	f.scheduler.ConfigureHW()

	f.scheduler.ScheduleHEFT()
}

func (f *Forge) Output() {
	f.scheduler.graph.OutputDotFile()
}
