package forge

//import "fmt"

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
		// Evaluate the graph with random inputs and set the outputs as golden
		g.EvaluateGolden(1)

		g.SimplifyArithmetic()
		g.EliminateDuplicatedOperation()
		g.MaximizeParallelism()
		g.DeleteUnusedNodes()

		// Evaluate the graph again using the same inputs and compare with the golden outputs
		g.EvaluateCompare()

		g.Analyze()

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

	f.scheduler.Schedule()
}

func (f *Forge) Output() {
	f.scheduler.graph.OutputDotFile()
}
