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

func printUint2(input Input) {
	val := 0
	select {
	case <-input.Fields["0"].Fields["0"].Unit:
	case <-input.Fields["0"].Fields["1"].Unit:
		val = 1
	}
	select {
	case <-input.Fields["1"].Fields["0"].Unit:
	case <-input.Fields["1"].Fields["1"].Unit:
		val += 2
	}
	fmt.Println(val)
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
	//ifile, err := os.Open("src/hello.don")
	//if err != nil {
	//	panic(err)
	//}
	ifile := os.Stdin

	hopefulInputType := MakeNStructType(2)
	hopefulInputType.Fields["0"] = types.Uint8Type
	hopefulInputType.Fields["1"] = types.Uint8Type

	//hopefulInputType := MakeNStructType(2)
	//hopefulInputType.Fields["0"] = types.BitType
	//hopefulInputType.Fields["1"] = types.BitType

	//hopefulOutputType := MakeNStructType(2)
	//hopefulOutputType.Fields["0"] = types.BitType
	//hopefulOutputType.Fields["1"] = types.BitType

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
