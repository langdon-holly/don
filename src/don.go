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

func checkTypes(com Com, inputType, outputType *DType, hopefulInputType, hopefulOutputType DType) {
	notUnderdefined := com.Types(inputType, outputType)

	if !notUnderdefined {
		fmt.Println("Input type:")
		fmt.Println(*inputType)
		fmt.Println("Output type:")
		fmt.Println(*outputType)
		panic("Underdefined types")
	}

	if !inputType.Equal(hopefulInputType) {
		fmt.Println("Input type:")
		fmt.Println(*inputType)
		panic("Bad input type")
	}
	if !outputType.Equal(hopefulOutputType) {
		fmt.Println("Output type:")
		fmt.Println(*outputType)
		panic("Bad output type")
	}
}

func main() {
	ifile := os.Stdin

	hopefulInputType := MakeNStructType(2)
	hopefulInputType.Fields["0"] = types.Uint8Type
	hopefulInputType.Fields["1"] = types.Uint8Type

	hopefulOutputType := types.Uint9Type

	com := syntax.ParseTop(ifile).ToCom(syntax.DefContext)

	var inputType, outputType DType
	inputType = hopefulInputType
	checkTypes(com, &inputType, &outputType, hopefulInputType, hopefulOutputType)

	inputR, input := MakeIO(inputType)
	output, outputW := MakeIO(outputType)
	go com.Run(inputType, outputType, inputR, outputW)

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
