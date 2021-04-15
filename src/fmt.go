package main

import (
	"fmt"
	"os"
)

import "don/syntax"

func main() { fmt.Println(syntax.TopString(syntax.ParseTop(os.Stdin))) }
