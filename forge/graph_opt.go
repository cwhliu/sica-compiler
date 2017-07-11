package forge

import "fmt"

func (g *Graph) optDeleteInternalNodes() {
  for name, node := range g.internalNodes {
    if node.numFanins() != 1 {
      fmt.Println("optimizer error - internal nodes should have single fanin")
      return
    }

    fi, _ := node.fanin(0)
    fi.removeFanout(node)

    for _, fo := range node.fanouts {
      fi.addFanout(fo)
      fo.replaceFanin(node, fi)
    }

    g.deleteNodeByName(name)
  }
}
