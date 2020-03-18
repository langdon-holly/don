package main

import (
	"fmt"
	"os"
)

import (
	. "don/core"
	"don/extra"
	"don/syntax"
	//"don/types"
)

func printTrit(input Input) {
	select {
	case <-input.Struct["0"].Unit:
		fmt.Println("0")
	case <-input.Struct["1"].Unit:
		fmt.Println("1")
	case <-input.Struct["2"].Unit:
		fmt.Println("2")
	}
}

func main() {
	ifile, err := os.Open("src/hello.don")
	if err != nil {
		panic(err)
	}

	inputTypeFields := make(map[string]DType, 2)
	inputTypeFields["inc"] = UnitType
	inputTypeFields["dec"] = UnitType
	inputType := MakeStructType(inputTypeFields)

	com := syntax.ParseTop(ifile)[0][0].ToCom()

	input, output, quit := extra.Run(com, inputType)
	defer close(quit)

	printTrit(output)
	input.Struct["dec"].WriteUnit()
	printTrit(output)
	input.Struct["dec"].WriteUnit()
	printTrit(output)
	input.Struct["inc"].WriteUnit()
	printTrit(output)
	input.Struct["dec"].WriteUnit()
	printTrit(output)
	input.Struct["inc"].WriteUnit()
	printTrit(output)
	input.Struct["inc"].WriteUnit()
	printTrit(output)
	input.Struct["inc"].WriteUnit()
	printTrit(output)
}
