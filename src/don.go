package main

import (
	"fmt"
	"os"
)

import (
	"don/extra"
	"don/syntax"
	"don/types"
)

func main() {
	ifile, err := os.Open("src/hello.don")
	if err != nil {
		panic(err)
	}

	com := syntax.ParseTop(ifile)[0][0].ToCom()

	input, output, quit := extra.Run(com, types.BoolType)
	defer close(quit)

	fmt.Println(types.ReadBool(output))

	types.WriteBool(input, true)
	fmt.Println(types.ReadBool(output))

	types.WriteBool(input, false)
	fmt.Println(types.ReadBool(output))
}
