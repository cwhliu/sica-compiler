package forge

type nodeCategory int

const (
	input nodeCategory = iota
	output
	internal
	operation
	constant
)

type nodeOperation int

const (
	equal nodeOperation = iota
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
