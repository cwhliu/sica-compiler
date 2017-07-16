package forge

import "fmt"

/*
Forge is the main compiler instance.
*/
type Forge struct {
	parser Parser
}

/*
Parse invokes the parser to parse the C++ source file and build a graph for it.
*/
func (f *Forge) Parse(filename string) error {
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

		g.OutputDotFile()

		return nil
	}
}
