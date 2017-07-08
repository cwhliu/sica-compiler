package forge

import (
	"fmt"

	"github.com/go-clang/v3.9/clang"
)

type parser struct {
	graph *graph
}

// Parse a MEX cc file and build a corresponding graph
func (p *parser) parse() (*graph, error) {
	p.graph = &graph{}

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

	cursor := tu.TranslationUnitCursor()

	cursor.Visit(func(cursor, parent clang.Cursor) clang.ChildVisitResult {
		//fmt.Println(cursor.Kind().Spelling())
		return clang.ChildVisit_Recurse
	})

	return p.graph, nil
}
