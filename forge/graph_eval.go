package forge

import (
	"fmt"
	"math"
	"math/rand"
)

var GraphPIs []map[string]float64
var GraphPOs []map[string]float64

/*
EvaluateGolden creates a number of sets of random input values, evaluates the
graph using these input values, and records the output values as the golden result.
*/
func (g *Graph) EvaluateGolden(numSets int) {
	g.Levelize()

	GraphPIs = make([]map[string]float64, numSets)
	GraphPOs = make([]map[string]float64, numSets)

	for set := 0; set < numSets; set++ {
		GraphPIs[set] = make(map[string]float64, g.NumInputNodes())
		GraphPOs[set] = make(map[string]float64, g.NumOutputNodes())

		for name, node := range g.inputNodes {
			GraphPIs[set][name] = rand.Float64()
			node.value = GraphPIs[set][name]
		}

		g.Eval()

		for name, node := range g.outputNodes {
			GraphPOs[set][name] = node.value
		}
	}
}

/*
EvaluateCompare uses the store input values to evaluate the graph, and compares
the output values against the golden result.

This function is used to verify that a graph retains the same functionality after
some graph transformation.
*/
func (g *Graph) EvaluateCompare() {
	g.Levelize()

	for set := 0; set < len(GraphPIs); set++ {
		for name, node := range g.inputNodes {
			node.value = GraphPIs[set][name]
		}

		g.Eval()

		for name, node := range g.outputNodes {
			result := node.value
			golden := GraphPOs[set][name]

			diffAbs := math.Abs(result - golden)
			diffRel := math.Abs(diffAbs / golden)

			// Mismatch if the absolute value differs by 0.01 and the relative value
			// differs by 1%
			if diffAbs > 0.01 && diffRel > 0.01 {
				fmt.Printf("Mismatch, set %d, node %s, %f != %f\n",
					set, node.name, result, golden)
			}
		}
	}
}
