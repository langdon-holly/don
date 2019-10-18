package main

import "fmt"

import (
	"don/coms"
	. "don/core"
	"don/extra"
)

func main() {
	com := coms.GenPipe([]GenCom{coms.GenI{}, coms.GenSplit{}, coms.GenAnd{}}).Com(BoolType)

	input, output, quit := extra.Run(com)
	defer close(quit)

	WriteBool(input, true)
	fmt.Println(ReadBool(output))

	WriteBool(input, true)
	fmt.Println(ReadBool(output))

	WriteBool(input, false)
	fmt.Println(ReadBool(output))
}
