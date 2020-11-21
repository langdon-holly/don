package coms

import . "don/core"

type UnitCom struct{}

func (UnitCom) Instantiate() ComInstance {
	ui := unitInstance(UnitType)
	return &ui
}

type unitInstance DType

func (ui *unitInstance) InputType() *DType  { return (*DType)(ui) }
func (ui *unitInstance) OutputType() *DType { return (*DType)(ui) }
func (unitInstance) Types() (underdefined Error) {
	return
}
func (ui unitInstance) Run(input Input, output Output) {
	RunI(DType(ui), input, output)
}
