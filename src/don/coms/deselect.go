package coms

import . "don/core"

type DeselectCom string

func (gd DeselectCom) OutputType(inputType PartialType) (ret PartialType) {
	ret.P = true
	ret.Tag = StructTypeTag

	ret.Fields = make(map[string]PartialType, 1)
	ret.Fields[string(gd)] = inputType

	return
}

func (gd DeselectCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	RunI(inputType, input, output.Struct[string(gd)], quit)
}
