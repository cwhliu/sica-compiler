package forge

import "fmt"

type Forge struct {
	Parser
}

func (f *Forge) Parse(fname string) error {
	if g, err := f.parse(fname); err != nil {
		return err
	} else {
    g.optDeleteInternalNodes()

		fmt.Printf("Graph has %d nodes\n", g.numAllNodes())

		g.outputDotFile()

		return nil
	}
}
