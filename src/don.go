package main

import "fmt"

import (
	"don/coms"
	. "don/core"
	"don/extra"
	"don/types"
)

func main() {
	com := coms.Pipe([]Com{coms.ICom{}, coms.SplitCom([]string{"hello", "hi"}), coms.AndCom{}})

	input, output, quit := extra.Run(com, types.BoolType)
	defer close(quit)

	types.WriteBool(input, true)
	fmt.Println(types.ReadBool(output))

	types.WriteBool(input, true)
	fmt.Println(types.ReadBool(output))

	types.WriteBool(input, false)
	fmt.Println(types.ReadBool(output))
}
