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

func printUint8(input Input) {
	fmt.Println(types.ReadUint8(input))
}

func checkTypes(com Com, inputType, outputType *DType, hopefulInputType, hopefulOutputType DType) {
	overdefined, notUnderdefined := com.Types(inputType, outputType)

	if overdefined != nil {
		for _, hm := range overdefined {
			fmt.Println(hm)
		}
		panic("Overdefined types")
	} else if !notUnderdefined {
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
	hopefulInputType.Fields["a"] = types.Uint8Type
	hopefulInputType.Fields["b"] = types.Uint8Type

	hopefulOutputType := types.Uint8Type

	com := syntax.ParseTop(ifile).ToCom(syntax.DefContext)

	var inputType, outputType DType
	inputType = hopefulInputType
	checkTypes(com, &inputType, &outputType, hopefulInputType, hopefulOutputType)

	inputR, input := MakeIO(inputType)
	output, outputW := MakeIO(outputType)
	go com.Run(inputType, outputType, inputR, outputW)

	types.WriteUint8(input.Fields["a"], 0)
	types.WriteUint8(input.Fields["b"], 0)
	printUint8(output)

	types.WriteUint8(input.Fields["a"], 2)
	types.WriteUint8(input.Fields["b"], 2)
	printUint8(output)

	types.WriteUint8(input.Fields["a"], 189)
	types.WriteUint8(input.Fields["b"], 55)
	printUint8(output)

	types.WriteUint8(input.Fields["a"], 255)
	types.WriteUint8(input.Fields["b"], 255)
	printUint8(output)
}
