package forge

type nodeKind int

const (
	input nodeKind = iota
	output
	internal
	operation
	constant
)

type nodeOp int

const (
	equal nodeOp = iota
	neg
	add
	sub
	mul
	div
	power
	sqrt
	abs
	exp
	log
	sin
	cos
	tan
	arcsin
	arccos
	arctan
	sinh
	cosh
	tanh
)

var validNodeOp = make(map[string]bool)

func init() {
}
