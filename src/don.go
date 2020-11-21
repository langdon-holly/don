package main

import (
	"fmt"
	"os"
)

import (
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

func main() {
	ifile := os.Stdin

	hopefulInputType := MakeNStructType(2)
	hopefulInputType.Fields["0"] = types.Uint8Type
	hopefulInputType.Fields["1"] = types.Uint8Type

	hopefulOutputType := types.Uint9Type

	comI := syntax.ParseTop(ifile).ToCom(syntax.DefContext).Instantiate()
	comI.InputType().Meets(hopefulInputType)

	checkTypes(comI, hopefulInputType, hopefulOutputType)

	inputR, input := MakeIO(*comI.InputType())
	output, outputW := MakeIO(*comI.OutputType())
	go comI.Run(inputR, outputW)

	types.WriteUint8(input.Fields["0"], 0)
	types.WriteUint8(input.Fields["1"], 0)
	printUint9(output)

	types.WriteUint8(input.Fields["0"], 2)
	types.WriteUint8(input.Fields["1"], 2)
	printUint9(output)

	types.WriteUint8(input.Fields["0"], 189)
	types.WriteUint8(input.Fields["1"], 55)
	printUint9(output)

	types.WriteUint8(input.Fields["0"], 255)
	types.WriteUint8(input.Fields["1"], 255)
	printUint9(output)
}
