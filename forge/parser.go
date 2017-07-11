package forge

import (
	"fmt"
	"strings"

	"github.com/cwhliu/go-clang/clang"
)

type parser struct {
	parserStack

	graph *graph
}

/*
Parse a MEX cc file and build a corresponding graph
*/
func (p *parser) parse(fname string) (*graph, error) {
	p.graph = createGraph()

	// Create a new index to store translation units
	//  arg1: exclude declarations from precompiled header
	//  arg2: display diagnostics
	idx := clang.NewIndex(1, 1)
	defer idx.Dispose()

	tuArgs := []string{
		"-DMATLAB_MEX_FILE",
		"-Itestdata/inc",
	}

	// Parse a given source file and its translation unit
	//  arg1: file name
	//  arg2: arguments
	//  arg3: unsaved files
	//  arg4: translation unit flags
	tu := idx.ParseTranslationUnit(fname, tuArgs, nil, 0)
	defer tu.Dispose()

	diag := tu.Diagnostics()

	// Check the translation unit is valid (source file exists) and there is
	// no problem parsing the source file
	if !tu.IsValid() || len(diag) > 0 {
		return nil, fmt.Errorf("problem parsing file %s", fname)
	}

	inTargetFunc := false

	cursor := tu.TranslationUnitCursor()

	// Recursively traverse the source file and build the graph
	buildOk := cursor.Visit(func(cursor, parent clang.Cursor) clang.ChildVisitResult {
		// Check if we are in the target function
		if cursor.Kind().Spelling() == "FunctionDecl" {
			if cursor.Spelling() == "output1" {
				inTargetFunc = true
			} else {
				inTargetFunc = false
			}
		}

		// If we are in the target function
		if inTargetFunc {
			switch cursor.Kind().Spelling() {
			case "DeclRefExpr":
				if !p.parseDeclRefExpr(cursor) {
					return clang.ChildVisit_Break
				}
			case "ArraySubscriptExpr":
				if !p.parseArraySubscriptExpr(cursor) {
					return clang.ChildVisit_Break
				}
			case "IntegerLiteral", "FloatingLiteral":
				if !p.parseLiteral(cursor) {
					return clang.ChildVisit_Break
				}
			case "UnaryOperator", "BinaryOperator":
				if !p.parseOperator(cursor) {
					return clang.ChildVisit_Break
				}
			}
		}

		return clang.ChildVisit_Recurse
	})

	// Source file traversal failed
	if !buildOk {
		return nil, fmt.Errorf("problem building graph for %s", fname)
	}

	p.graph.finalize()
	fmt.Printf("%s %d ", fname, len(p.tokens))

	return p.graph, nil
}

// Parser sub-functions for specific AST nodes
// -----------------------------------------------------------------------------

/*
Parse reference to a declared expression
*/
func (p *parser) parseDeclRefExpr(cursor clang.Cursor) bool {
	cursorType := cursor.Type().Spelling()

	// It's a reference to a function
	if strings.Contains(cursorType, "(") {
		// Extract the parameter list
		start := strings.Index(cursorType, "(") + 1
		stop := strings.Index(cursorType, ")")

		cursorType = cursorType[start:stop]

		// Find out how many parameters this function has
		numParms := len(strings.Split(cursorType, ","))

		// Only support functions with up to 2 parameters
		if numParms > 2 {
			fmt.Printf("parse error - support functions with up to 2 parameters")

			return false
		}

		// Push a non-leaf FUN token to the stack
		p.pushNonLeafToken("FUN"+cursor.Spelling(), numParms)
	} else { // It's a reference to a variable
		// Push a leaf VAR token to the stack
		p.pushLeafToken("VAR" + cursor.Spelling())
		// Process the stack whenever a leaf token is pushed
		p.processStack()
	}

	return true
}

/*
Parse array expression
*/
func (p *parser) parseArraySubscriptExpr(cursor clang.Cursor) bool {
	// Push a non-leaf ARR token to the stack
	p.pushNonLeafToken("ARR", 2)

	return true
}

/*
Parse literal
*/
func (p *parser) parseLiteral(cursor clang.Cursor) bool {
	switch cursor.Kind().Spelling() {
	case "IntegerLiteral":
		p.pushLeafToken("CON" + cursor.LiteralSpelling())
		// Process the stack whenever a leaf token is pushed
		p.processStack()
	case "FloatingLiteral":
		// Remove trailing zeros and decimal point, for example
		// 1.200 becomes 1,2 and 3.0 become 3
		p.pushLeafToken("CON" + strings.TrimRight(cursor.LiteralSpelling(), "0."))
		// Process the stack whenever a leaf token is pushed
		p.processStack()
	}

	return true
}

/*
Parse operator
*/
func (p *parser) parseOperator(cursor clang.Cursor) bool {
	switch cursor.Kind().Spelling() {
	case "UnaryOperator":
		p.pushNonLeafToken("UOP"+cursor.OperatorSpelling(), 1)
	case "BinaryOperator":
		p.pushNonLeafToken("BOP"+cursor.OperatorSpelling(), 2)
	}

	return true
}

// -----------------------------------------------------------------------------

func (p *parser) processStack() {
	// Loop whenever a token is ready to be popped
	for p.tokenReady() {
		// Pop the token and its arguments from the stack
		token, args := p.popToken()

		// Token/argument type represented by the first 3 characters
		// Token/argument value represented by the rest of the characters

		switch token[0:3] {
		default:
		case "ARR":
			operand := p.graph.getNodeByName("ARR" + args[0][3:] + "[" + args[1][3:] + "]")

			p.pushLeafToken(operand.name)
		case "BOP":
			opcode := token[3:]

			lOperand := p.graph.getNodeByName(args[0])
			rOperand := p.graph.getNodeByName(args[1])

			if opcode == "=" {
				lOperand.receive(rOperand)
			} else {
				opNode := p.graph.addOperationNode(opcode)
				opNode.receive(lOperand)
				opNode.receive(rOperand)

				p.pushLeafToken(opNode.name)
			}
		case "UOP":
			opcode := token[3:]

			if opcode == "-" {
				operand := p.graph.getNodeByName(args[0])

				opNode := p.graph.addOperationNode(opcode)
				opNode.receive(operand)

				p.pushLeafToken(opNode.name)
			} else {
				fmt.Println("parser error - only support - unary operator")
			}
		case "FUN":
			funcName := strings.ToLower(token[3:])
			numParms := len(args)

			if numParms == 2 {
				operand1 := p.graph.getNodeByName(args[0])
				operand2 := p.graph.getNodeByName(args[1])

				opNode := p.graph.addOperationNode(funcName)
				opNode.receive(operand1)
				opNode.receive(operand2)

				p.pushLeafToken(opNode.name)
			} else {
				operand := p.graph.getNodeByName(args[0])

				opNode := p.graph.addOperationNode(funcName)
				opNode.receive(operand)

				p.pushLeafToken(opNode.name)
			}
		}
	}
}
