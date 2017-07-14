package forge

import "fmt"

type Forge struct {
	parser Parser
}

func (f *Forge) Parse(fname string) error {
	if g, err := f.parser.Parse(fname); err != nil {
		return err
	} else {
		g.OptimizeValueNumbering()
		g.OptimizeTreeHeight()

		fmt.Printf("Graph has %d operation nodes\n", g.NumOperationNodes())

		g.OutputDotFile()

		return nil
	}
}
