package main

import (
	"fmt"
	"os"
)

import "don/syntax"

func main() { fmt.Println(syntax.Parse(os.Stdin)) }
