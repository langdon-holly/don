package main

import (
	"fmt"
	"os"
)

import (
	. "don/core"
	"don/extra"
	"don/syntax"
	"don/types"
)

func printUint9(input Input) {
	val := 0
	for i, fieldName := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"} {
		select {
		case <-input.Struct[fieldName].Struct["0"].Unit:
		case <-input.Struct[fieldName].Struct["1"].Unit:
			val += 1 << i
		}
	}
	fmt.Println(val)
}

func main() {
	ifile, err := os.Open("src/hello.don")
	if err != nil {
		panic(err)
	}

	inputTypeFields := make(map[string]DType, 2)
	inputTypeFields["a"] = types.Uint8Type
	inputTypeFields["b"] = types.Uint8Type
	inputType := MakeStructType(inputTypeFields)

	com := syntax.ParseTop(ifile).ToCom(syntax.DefContext)

	input, output, quit := extra.Run(com, inputType)
	defer close(quit)

	types.WriteUint8(input.Struct["a"], 0)
	types.WriteUint8(input.Struct["b"], 0)
	printUint9(output)

	types.WriteUint8(input.Struct["a"], 2)
	types.WriteUint8(input.Struct["b"], 2)
	printUint9(output)

	types.WriteUint8(input.Struct["a"], 189)
	types.WriteUint8(input.Struct["b"], 55)
	printUint9(output)

	types.WriteUint8(input.Struct["a"], 255)
	types.WriteUint8(input.Struct["b"], 255)
	printUint9(output)
}
