package forge

import "fmt"

type Forge struct {
	parser Parser
}

func (f *Forge) Parse(fname string) error {
	if g, err := f.parser.Parse(fname); err != nil {
		return err
	} else {
		g.EvaluateGolden(1)

		g.OptimizeValueNumbering()
		g.OptimizeTreeHeight()

		g.EvaluateCompare()

		fmt.Printf(" %d operation nodes, %d levels\n",
			g.NumOperationNodes(), g.Levelize())

		g.OutputDotFile()

		return nil
	}
}
