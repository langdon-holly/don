package main

import (
	"fmt"
	"os"
)

import (
	"don/coms"
	. "don/core"
	"don/syntax"
	"don/types"
)

func printUint9(rMap ReadMap) { fmt.Println(types.ReadUint9(rMap)) }

func checkTypes(com Com, hopefulInputType, hopefulOutputType DType) {
	if underdefined := com.Underdefined(); underdefined != nil {
		fmt.Println(underdefined)
		panic("Underdefined types")
	} else if !com.InputType().Equal(hopefulInputType) {
		fmt.Println("Input type:")
		fmt.Println(com.InputType())
		panic("Bad input type")
	} else if !com.OutputType().Equal(hopefulOutputType) {
		fmt.Println("Output type:")
		fmt.Println(com.OutputType())
		panic("Bad output type")
	}
}

func runWithInputs(tc TypedCom, arg0, arg1 int) {
	wMap, rMap := tc.Run()
	types.WriteUint8(wMap.Fields["0"], arg0)
	types.WriteUint8(wMap.Fields["1"], arg1)
	printUint9(rMap)
}

func main() {
	ifile := os.Stdin

	hopefulInputType := MakeNFieldsType(2)
	hopefulInputType.Fields["0"] = types.Uint8Type
	hopefulInputType.Fields["1"] = types.Uint8Type

	hopefulOutputType := types.Uint9Type

	com := coms.Eval(syntax.ParseTop(ifile), coms.DefContext).Com().MeetTypes(
		hopefulInputType,
		UnknownType,
	)

	checkTypes(com, hopefulInputType, hopefulOutputType)

	tc := MakeTypedCom(com)
	tc.Determinate()

	runWithInputs(tc, 0, 0)
	runWithInputs(tc, 2, 2)
	runWithInputs(tc, 189, 55)
	runWithInputs(tc, 255, 255)
}
