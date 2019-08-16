package main

import "fmt"

import (
	"don/coms"
	. "don/core"
	"don/extra"
)

func main() {
	com := coms.GenPipe([]GenCom{coms.GenI, coms.GenSplit, coms.GenAnd})(BoolType)

	inputI, inputO := extra.MakeIOChans(com.InputType())
	outputI, outputO := extra.MakeIOChans(com.OutputType())

	quit := make(chan struct{})
	defer close(quit)

	go com.Run(inputI, outputO, quit)

	WriteBool(inputO, true)
	fmt.Println(ReadBool(outputI))

	WriteBool(inputO, true)
	fmt.Println(ReadBool(outputI))

	WriteBool(inputO, false)
	fmt.Println(ReadBool(outputI))
}
