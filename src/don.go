package main

import "fmt"

import (
	"don/coms"
	. "don/core"
	"don/extra"
	"don/types"
)

func main() {
	com := coms.Pipe([]Com{
		coms.ICom{},
		coms.SplitCom([]string{"0", "1"}),
		coms.ProdCom{},
		coms.Deselect("l"),
		coms.SelectCom("l")})

	input, output, quit := extra.Run(com, types.BoolType)
	defer close(quit)

	types.WriteBool(input, true)
	<-output.Struct["true"].Struct["true"].Unit
	fmt.Println(":[true] :[true]")

	types.WriteBool(input, false)
	<-output.Struct["false"].Struct["false"].Unit
	fmt.Println(":[false] :[false]")
}
