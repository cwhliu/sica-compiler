package main

import (
	"fmt"
	"os"

	"github.com/cwhliu/sica-compiler/forge"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("\nUsage: %s file_name\n\n", os.Args[0])
		return
	}

	f := forge.Forge{}

	if err := f.BuildGraph(os.Args[1]); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	f.ScheduleGraph()
}
