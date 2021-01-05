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

func printUint9(input Input) {
	fmt.Println(types.ReadUint9(input))
}

func checkTypes(comI ComInstance, hopefulInputType, hopefulOutputType DType) {
	comI.Types()
	if underdefined := comI.Underdefined(); underdefined != nil {
		fmt.Println(underdefined)
		panic("Underdefined types")
	} else if !comI.InputType().Equal(hopefulInputType) {
		fmt.Println("Input type:")
		fmt.Println(*comI.InputType())
		panic("Bad input type")
	} else if !comI.OutputType().Equal(hopefulOutputType) {
		fmt.Println("Output type:")
		fmt.Println(*comI.OutputType())
		panic("Bad output type")
	}
}

func runWithInputs(comI ComInstance, arg0, arg1 int) {
	inputR, input := MakeIO(*comI.InputType())
	output, outputW := MakeIO(*comI.OutputType())
	go comI.Run(inputR, outputW)

	types.WriteUint8(input.Fields["0"], arg0)
	types.WriteUint8(input.Fields["1"], arg1)
	printUint9(output)
}

func main() {
	ifile := os.Stdin

	hopefulInputType := MakeNFieldsType(2)
	hopefulInputType.Fields["0"] = types.Uint8Type
	hopefulInputType.Fields["1"] = types.Uint8Type

	hopefulOutputType := types.Uint9Type

	comI := coms.Eval(syntax.ParseTop(ifile), coms.DefContext).Com().Instantiate()
	comI.InputType().Meets(hopefulInputType)

	checkTypes(comI, hopefulInputType, hopefulOutputType)

	runWithInputs(comI, 0, 0)
	runWithInputs(comI, 2, 2)
	runWithInputs(comI, 189, 55)
	runWithInputs(comI, 255, 255)
}
