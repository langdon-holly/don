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

func printUint9At(rMap ReadMap, path []string) {
	fmt.Println(types.ReadUint9At(rMap, path))
}

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

func runWithInputs(tc TypedCom, arg0, arg1, arg2, arg3 int) {
	wMap, rMap := tc.Run()

	types.WriteUint8At(wMap.Fields["0"], arg0, []string{"0"})
	types.WriteUint8At(wMap.Fields["1"], arg1, []string{"0"})
	printUint9At(rMap, []string{"0"})

	types.WriteUint8At(wMap.Fields["0"], arg2, []string{"1"})
	types.WriteUint8At(wMap.Fields["1"], arg3, []string{"1"})
	printUint9At(rMap, []string{"1"})
}

func main() {
	ifile := os.Stdin

	singleInputType := MakeNFieldsType(2)
	singleInputType.Fields["0"] = types.Uint8Type
	singleInputType.Fields["1"] = types.Uint8Type
	hopefulInputType := singleInputType.AgainstPath([]string{"0"})
	hopefulInputType.Joins(singleInputType.AgainstPath([]string{"1"}))

	hopefulOutputType := types.Uint9Type.AgainstPath([]string{"0"})
	hopefulOutputType.Joins(types.Uint9Type.AgainstPath([]string{"1"}))

	com := coms.Eval(syntax.ParseTop(ifile), coms.DefContext).Com().MeetTypes(
		hopefulInputType,
		UnknownType,
	)

	checkTypes(com, hopefulInputType, hopefulOutputType)

	tc := MakeTypedCom(com)
	tc.Determinate()

	runWithInputs(tc, 0, 0, 2, 2)
	runWithInputs(tc, 189, 55, 255, 255)
}
