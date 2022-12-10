package main

import (
	"fmt"
	"os"
)

import (
	"don/rel"
)

func main() {
	fmt.Println(rel.TypePtrType(rel.EvalFile(os.Args[1]).Rel().Type()))
}
