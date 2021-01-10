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

func checkTypes(com Com, hopefulInputType, hopefulOutputType DType) {
	com = com.Types()
	if underdefined := com.Underdefined(); underdefined != nil {
		fmt.Println(underdefined)
		panic("Underdefined types")
	} else if !com.InputType().Equal(hopefulInputType) {
		fmt.Println("Input type:")
		fmt.Println(*com.InputType())
		panic("Bad input type")
	} else if !com.OutputType().Equal(hopefulOutputType) {
		fmt.Println("Output type:")
		fmt.Println(*com.OutputType())
		panic("Bad output type")
	}
}

func runWithInputs(com Com, arg0, arg1 int) {
	inputR, input := MakeIO(*com.InputType())
	output, outputW := MakeIO(*com.OutputType())
	go com.Run(inputR, outputW)

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

	com := coms.Eval(syntax.ParseTop(ifile), coms.DefContext).Com()
	com.InputType().Meets(hopefulInputType)

	checkTypes(com, hopefulInputType, hopefulOutputType)

	runWithInputs(com, 0, 0)
	runWithInputs(com, 2, 2)
	runWithInputs(com, 189, 55)
	runWithInputs(com, 255, 255)
}
