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

func (g *Graph) OptValueNumbering() {
	vnMap := make(map[string]*Node)

	for name, node := range g.operationNodes {
		vnKey := NodeOpStringLUT[node.op]
		for _, fi := range node.fanins {
			vnKey += fi.name
		}

		if vnNode, exist := vnMap[vnKey]; !exist {
			//fmt.Println("new", name, vnKey)

			vnMap[vnKey] = node
		} else {
			//fmt.Println("Reuse", name, vnKey)

			for _, fi := range node.fanins {
				fi.RemoveFanout(node)
			}

			for _, fo := range node.fanouts {
				vnNode.AddFanout(fo)
				fo.ReplaceFanin(node, vnNode)
			}

			g.DeleteNodeByName(name)
		}
	}
}
