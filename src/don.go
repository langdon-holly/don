package main

import (
	"fmt"
)

import (
	"don/coms"
	. "don/core"
	"don/extra"
	"don/types"
)

func main() {
	inputTypeFields := make(map[string]DType, 2)
	inputTypeFields["0"] = types.BoolType
	inputTypeFields["1"] = types.BoolType
	inputType := MakeStructType(inputTypeFields)
	com := coms.And

	input, output, quit := extra.Run(com, inputType)
	defer close(quit)

	types.WriteBool(input.Struct["0"], false)
	types.WriteBool(input.Struct["1"], false)
	fmt.Println(types.ReadBool(output))

	types.WriteBool(input.Struct["0"], false)
	types.WriteBool(input.Struct["1"], true)
	fmt.Println(types.ReadBool(output))

	types.WriteBool(input.Struct["0"], true)
	types.WriteBool(input.Struct["1"], false)
	fmt.Println(types.ReadBool(output))

	types.WriteBool(input.Struct["0"], true)
	types.WriteBool(input.Struct["1"], true)
	fmt.Println(types.ReadBool(output))
}
