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
	com := syntax.ParseTop(os.Stdin)[0][0].ToCom()

	input, output, quit := extra.Run(com, types.BoolType)
	defer close(quit)

	types.WriteBool(input, true)
	fmt.Println(types.ReadBool(output))

	types.WriteBool(input, false)
	fmt.Println(types.ReadBool(output))
}
