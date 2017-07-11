package forge

import "fmt"

type Forge struct {
	parser
}

func (f *Forge) Parse(fname string) error {
	if g, err := f.parse(fname); err != nil {
		return err
	} else {
		fmt.Printf("Graph has %d nodes\n", g.numAllNodes())

		g.outputDotFile()

		return nil
	}
}
