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
func (f *Forge) BuildGraph(filename, postfix string) {
	g, _ := f.parser.Parse(filename)

	// Evaluate the graph with random inputs and set the outputs as golden
	//g.EvaluateGolden(1)

	g.SimplifyArithmetic()
	g.EliminateDuplicatedOperation()
	g.MaximizeParallelism()
	g.DeleteUnusedNodes()

	// Evaluate the graph again using the same inputs and compare with the golden outputs
	//g.EvaluateCompare()

	//g.Analyze()

	if postfix != "" {
		g.AddPostfix(postfix)
	}

	// Pass the graph to the scheduler
	if f.scheduler.graph == nil {
		f.scheduler.graph = g
	} else {
		if f.scheduler.mergedGraph == nil {
			f.scheduler.mergedGraph = f.scheduler.graph
			f.scheduler.mergedGraph.Merge(g)
		} else {
			f.scheduler.mergedGraph.Merge(g)
		}

		f.scheduler.graph = g
	}
}

/*
ScheduleGraph schedules the operations in the graph onto the hardware accelerator.
*/
func (f *Forge) ScheduleGraph() {
	if f.scheduler.processor == nil {
		f.scheduler.processor = CreateProcessor()
	}

	f.scheduler.ScheduleHeuristic()
}

func (f *Forge) Output() {
	if f.scheduler.mergedGraph == nil {
		f.scheduler.graph.OutputDotFile()
	} else {
		f.scheduler.mergedGraph.OutputDotFile()
	}
}
