package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/cwhliu/sica-compiler/forge"
)

func main() {
	f := forge.Forge{}

	if len(os.Args) == 2 {
		f.BuildGraph(os.Args[1], "")
		f.ScheduleGraph()
	} else if len(os.Args) > 2 {
		for g := 1; g < len(os.Args); g++ {
			f.BuildGraph(os.Args[g], strconv.FormatInt(int64(g), 10))
			f.ScheduleGraph()
		}
	} else {
		fmt.Printf("\nUsage: %s file_name [file_names]\n\n", os.Args[0])
		return
	}

	f.Output()
}
