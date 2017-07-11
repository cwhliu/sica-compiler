package forge

import "fmt"

func (g *Graph) OptDeleteInternalNodes() {
	for name, node := range g.internalNodes {
		if node.NumFanins() != 1 {
			fmt.Println("optimizer error - internal nodes should have single fanin")
			return
		}

		fi, _ := node.Fanin(0)
		fi.RemoveFanout(node)

		for _, fo := range node.fanouts {
			fi.AddFanout(fo)
			fo.ReplaceFanin(node, fi)
		}

		g.DeleteNodeByName(name)
	}
}
