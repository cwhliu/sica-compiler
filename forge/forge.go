package forge

import "fmt"

type Forge struct {
	parser Parser
}

func (f *Forge) Parse(fname string) error {
	if g, err := f.parser.Parse(fname); err != nil {
		return err
	} else {
		g.OptimizeInternalNodes()
		g.OptimizeValueNumbering()
		g.OptimizeTreeHeight()

		fmt.Printf("Graph has %d nodes\n", g.NumAllNodes())

		g.OutputDotFile()

		return nil
	}
}
