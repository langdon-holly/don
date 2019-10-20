package coms

import . "don/core"
import "don/extra"

type SinkCom struct{}

func (SinkCom) OutputType(inputType PartialType) PartialType {
	return PartialType{
		P:      true,
		Tag:    StructTypeTag,
		Fields: make(map[string]PartialType, 0)}
}

func (SinkCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	_, newOutput := extra.MakeIOChans(inputType, 0)
	RunI(inputType, input, newOutput, quit)
}
