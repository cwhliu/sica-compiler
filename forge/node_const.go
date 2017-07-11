package forge

type nodeKind int

const (
	undetermined nodeKind = iota
	input
	output
	internal
	operation
	constant
)

type nodeOp int

const (
	nop nodeOp = iota
	equal
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

var nodeOpLUT = make(map[string]nodeOp)
var nodeOpStringLUT = make(map[nodeOp]string)

func init() {
	nodeOpLUT[""] = nop
	nodeOpLUT["="] = equal
	nodeOpLUT["+"] = add
	nodeOpLUT["-"] = sub
	nodeOpLUT["*"] = mul
	nodeOpLUT["/"] = div
	nodeOpLUT["power"] = power
	nodeOpLUT["sqrt"] = sqrt
	nodeOpLUT["abs"] = abs
	nodeOpLUT["exp"] = exp
	nodeOpLUT["log"] = log
	nodeOpLUT["sin"] = sin
	nodeOpLUT["cos"] = cos
	nodeOpLUT["tan"] = tan
	nodeOpLUT["arcsin"] = arcsin
	nodeOpLUT["arccos"] = arccos
	nodeOpLUT["arctan"] = arctan
	nodeOpLUT["sinh"] = sinh
	nodeOpLUT["cosh"] = cosh
	nodeOpLUT["tanh"] = tanh

	nodeOpStringLUT[nop] = ""
	nodeOpStringLUT[equal] = "="
	nodeOpStringLUT[add] = "+"
	nodeOpStringLUT[sub] = "-"
	nodeOpStringLUT[mul] = "*"
	nodeOpStringLUT[div] = "/"
	nodeOpStringLUT[power] = "power"
	nodeOpStringLUT[sqrt] = "sqrt"
	nodeOpStringLUT[abs] = "abs"
	nodeOpStringLUT[exp] = "exp"
	nodeOpStringLUT[log] = "log"
	nodeOpStringLUT[sin] = "sin"
	nodeOpStringLUT[cos] = "cos"
	nodeOpStringLUT[tan] = "tan"
	nodeOpStringLUT[arcsin] = "arcsin"
	nodeOpStringLUT[arccos] = "arccos"
	nodeOpStringLUT[arctan] = "arctan"
	nodeOpStringLUT[sinh] = "sinh"
	nodeOpStringLUT[cosh] = "cosh"
	nodeOpStringLUT[tanh] = "tanh"
}
