package main

import (
	"fmt"

	"github.com/cwhliu/sica-compiler/forge"
)

func main() {
	f := forge.Forge{}

	if err := f.Parse(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}
