package forge

import (
	"fmt"
	"strings"

	"github.com/go-clang/v3.9/clang"
)

type parser struct {
	parserStack

	graph *graph
}

// Parse a MEX cc file and build a corresponding graph
func (p *parser) parse() (*graph, error) {
	p.graph = createGraph()

	// Create a new index to store translation units
	//  arg1: exclude declarations from precompiled header
	//  arg2: display diagnostics
	idx := clang.NewIndex(1, 1)
	defer idx.Dispose()

	fname := "testdata/atlas/default/torque_LeftStance_interior.cc"

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
			case "ParmDecl":
				if !p.parseParmDecl(cursor) {
					return clang.ChildVisit_Break
				}
			case "VarDecl":
				if !p.parseVarDecl(cursor) {
					return clang.ChildVisit_Break
				}
			case "DeclRefExpr":
				if !p.parseDeclRefExpr(cursor) {
					return clang.ChildVisit_Break
				}
			case "ArraySubscriptExpr":
				if !p.parseArraySubscriptExpr(cursor) {
					return clang.ChildVisit_Break
				}
			case "BinaryOperator":
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

	return p.graph, nil
}

// Parser sub-functions for specific AST nodes
// -----------------------------------------------------------------------------

// Parse function parameter declaration
func (p *parser) parseParmDecl(cursor clang.Cursor) bool {
	// TODO
	return true
}

// Parse variable declaration
func (p *parser) parseVarDecl(cursor clang.Cursor) bool {
	return p.graph.addInternalNode(cursor.Spelling())
}

// Parse reference to a declared expression
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

		// Push a non-leaf FUN node onto stack
		p.pushNonLeaf("FUN"+cursor.Spelling(), numParms)
		// It's a reference to a variable
	} else {
		// Push a leaf VAR node onto stack
		p.pushLeaf("VAR" + cursor.Spelling())
	}

	return true
}

// Parse array expression
func (p *parser) parseArraySubscriptExpr(cursor clang.Cursor) bool {
	// Push a non-leaf ARR node onto stack
	p.pushNonLeaf("ARR", 2)

	return true
}

// Parse operator
func (p *parser) parseOperator(cursor clang.Cursor) bool {
	return true
}
