package main

import (
	"fmt"
	"os"
)

import "don/syntax"

func main() { fmt.Println(syntax.ParseTop(os.Stdin).TopString()) }
