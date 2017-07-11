package forge

import (
	"bufio"
	"fmt"
	"os"
)

func (g *Graph) OutputDotFile() {
	f, _ := os.Create("graph.dot")
	defer f.Close()

	w := bufio.NewWriter(f)

	w.WriteString("digraph demo {\n")
	w.WriteString("rankdir=TB\n")

	// Input nodes
	w.WriteString("{rank=min\n")
	for _, node := range g.inputNodes {
		label := node.name[3:]

		w.WriteString(fmt.Sprintf("\"%s\" ", node.name))
		w.WriteString(fmt.Sprintf("[shape=rect style=\"rounded,filled\""))
		w.WriteString(fmt.Sprintf(" fillcolor=deepskyblue label=\"%s\"]\n", label))
	}
	w.WriteString("}\n")

	// Output nodes
	w.WriteString("{rank=max\n")
	for _, node := range g.outputNodes {
		label := node.name[3:]

		w.WriteString(fmt.Sprintf("\"%s\" ", node.name))
		w.WriteString(fmt.Sprintf("[shape=rect style=\"rounded,filled\""))
		w.WriteString(fmt.Sprintf(" fillcolor=deepskyblue4 fontcolor=white label=\"%s\"]\n", label))
	}
	w.WriteString("}\n")

	// Constant nodes
	for _, node := range g.constantNodes {
		label := node.name[3:]

		w.WriteString(fmt.Sprintf("\"%s\" ", node.name))
		w.WriteString(fmt.Sprintf("[shape=plaintext label=\"%s\"]\n", label))
	}

	// Operation nodes
	for _, node := range g.operationNodes {
		opString, _ := NodeOpStringLUT[node.op]

		label := opString

		w.WriteString(fmt.Sprintf("\"%s\" ", node.name))
		w.WriteString(fmt.Sprintf("[shape=rect label=\"%s\"]\n", label))
	}

	// Edges
	for _, node := range g.allNodes {
		for i := 0; i < node.NumFanins(); i++ {
			fanin, _ := node.Fanin(i)

			w.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", fanin.name, node.name))
		}
	}

	w.WriteString("}\n")

	w.Flush()
}
