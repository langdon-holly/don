package main

import "fmt"

import (
	"don/coms"
	. "don/core"
	"don/extra"
)

func main() {
	com := coms.Pipe([]Com{coms.ICom{}, coms.SplitCom([]string{"hello", "hi"}), coms.AndCom{}})

	input, output, quit := extra.Run(com, BoolType)
	defer close(quit)

	WriteBool(input, true)
	fmt.Println(ReadBool(output))

	WriteBool(input, true)
	fmt.Println(ReadBool(output))

	WriteBool(input, false)
	fmt.Println(ReadBool(output))
}
