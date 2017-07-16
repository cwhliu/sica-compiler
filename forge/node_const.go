package forge

// -----------------------------------------------------------------------------

/*
NodeKind represents the type of a node.
*/
type NodeKind int

const (
	NodeKind_Undetermined NodeKind = iota
	NodeKind_Input
	NodeKind_Output
	NodeKind_Internal
	NodeKind_Operation
	NodeKind_Constant
)

// -----------------------------------------------------------------------------

// NodeOp represents the operation of a node.
type NodeOp int

const (
	NodeOp_Nop NodeOp = iota
	NodeOp_Equal
	NodeOp_Add
	NodeOp_Sub
	NodeOp_Mul
	NodeOp_Div
	NodeOp_Power
	NodeOp_Sqrt
	NodeOp_Abs
	NodeOp_Exp
	NodeOp_Log
	NodeOp_Sin
	NodeOp_Cos
	NodeOp_Tan
	NodeOp_Arcsin
	NodeOp_Arccos
	NodeOp_Arctan
	NodeOp_Sinh
	NodeOp_Cosh
	NodeOp_Tanh
)

// -----------------------------------------------------------------------------

// NodeOpLUT is a lookup table for converting a string to a NodeOp.
// For example, "+" -> NodeOp_Add
var NodeOpLUT = make(map[string]NodeOp)

// NodeOpStringLUT is a lookup table for converting a NodeOp to a string.
// For example, NodeOp_Add -> "+"
var NodeOpStringLUT = make(map[NodeOp]string)

func init() {
	// Commented operations are temporarily made unsupported
	NodeOpLUT[""] = NodeOp_Nop
	NodeOpLUT["="] = NodeOp_Equal
	NodeOpLUT["+"] = NodeOp_Add
	NodeOpLUT["-"] = NodeOp_Sub
	NodeOpLUT["*"] = NodeOp_Mul
	NodeOpLUT["/"] = NodeOp_Div
	NodeOpLUT["power"] = NodeOp_Power
	//NodeOpLUT["sqrt"] = NodeOp_Sqrt
	//NodeOpLUT["abs"] = NodeOp_Abs
	//NodeOpLUT["exp"] = NodeOp_Exp
	//NodeOpLUT["log"] = NodeOp_Log
	NodeOpLUT["sin"] = NodeOp_Sin
	NodeOpLUT["cos"] = NodeOp_Cos
	//NodeOpLUT["tan"] = NodeOp_Tan
	//NodeOpLUT["arcsin"] = NodeOp_Arcsin
	//NodeOpLUT["arccos"] = NodeOp_Arccos
	//NodeOpLUT["arctan"] = NodeOp_Arctan
	//NodeOpLUT["sinh"] = NodeOp_Sinh
	//NodeOpLUT["cosh"] = NodeOp_Cosh
	//NodeOpLUT["tanh"] = NodeOp_Tanh

	NodeOpStringLUT[NodeOp_Nop] = ""
	NodeOpStringLUT[NodeOp_Equal] = "="
	NodeOpStringLUT[NodeOp_Add] = "+"
	NodeOpStringLUT[NodeOp_Sub] = "-"
	NodeOpStringLUT[NodeOp_Mul] = "*"
	NodeOpStringLUT[NodeOp_Div] = "/"
	NodeOpStringLUT[NodeOp_Power] = "power"
	//NodeOpStringLUT[NodeOp_Sqrt] = "sqrt"
	//NodeOpStringLUT[NodeOp_Abs] = "abs"
	//NodeOpStringLUT[NodeOp_Exp] = "exp"
	//NodeOpStringLUT[NodeOp_Log] = "log"
	NodeOpStringLUT[NodeOp_Sin] = "sin"
	NodeOpStringLUT[NodeOp_Cos] = "cos"
	//NodeOpStringLUT[NodeOp_Tan] = "tan"
	//NodeOpStringLUT[NodeOp_Arcsin] = "arcsin"
	//NodeOpStringLUT[NodeOp_Arccos] = "arccos"
	//NodeOpStringLUT[NodeOp_Arctan] = "arctan"
	//NodeOpStringLUT[NodeOp_Sinh] = "sinh"
	//NodeOpStringLUT[NodeOp_Cosh] = "cosh"
	//NodeOpStringLUT[NodeOp_Tanh] = "tanh"
}
