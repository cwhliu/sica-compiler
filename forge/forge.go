package forge

import "fmt"

type Forge struct {
	parser
}

func (f *Forge) Parse() error {
	if g, err := f.parse(); err != nil {
		return err
	} else {
		fmt.Printf("Graph has %d nodes\n", g.numAllNodes())

		return nil
	}
}
