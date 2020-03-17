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

func main() {
	ifile, err := os.Open("src/hello.don")
	if err != nil {
		panic(err)
	}

	inputTypeFields := make(map[string]DType, 2)
	inputTypeFields["a"] = types.BoolType
	inputTypeFields["b"] = types.BoolType
	inputType := MakeStructType(inputTypeFields)

	com := syntax.ParseTop(ifile)[0][0].ToCom()

	input, output, quit := extra.Run(com, inputType)
	defer close(quit)

	types.WriteBool(input.Struct["a"], false)
	types.WriteBool(input.Struct["b"], false)
	fmt.Println(types.ReadBool(output))

	types.WriteBool(input.Struct["a"], false)
	types.WriteBool(input.Struct["b"], true)
	fmt.Println(types.ReadBool(output))

	types.WriteBool(input.Struct["a"], true)
	types.WriteBool(input.Struct["b"], false)
	fmt.Println(types.ReadBool(output))

	types.WriteBool(input.Struct["a"], true)
	types.WriteBool(input.Struct["b"], true)
	fmt.Println(types.ReadBool(output))
}
