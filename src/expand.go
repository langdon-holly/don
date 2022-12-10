package main

import (
	"fmt"
	"os"
)

import (
	"don/com"
)

func main() { fmt.Println(com.EvalFile(os.Args[1]).Com()) }
