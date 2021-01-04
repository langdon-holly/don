package main

import (
	"fmt"
	"os"
)

import "don/syntax"

func main() {
	fmt.Print(syntax.ParseTop(os.Stdin).Children[1].StringAtTop()[1:])
}
