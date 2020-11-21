package coms

import . "don/core"

type StructCom struct{}

func (StructCom) Instantiate() ComInstance {
	si := structInstance(StructType)
	return &si
}

type structInstance DType

func (si *structInstance) InputType() *DType  { return (*DType)(si) }
func (si *structInstance) OutputType() *DType { return (*DType)(si) }
func (si structInstance) Types()              {}
func (si structInstance) Underdefined() Error {
	return DType(si).Underdefined().Context("in struct")
}
func (si structInstance) Run(input Input, output Output) {
	RunI(DType(si), input, output)
}
