package coms

import . "don/core"

type UnitCom struct{}

func (UnitCom) Instantiate() ComInstance {
	ui := unitInstance(UnitType)
	return &ui
}

func (UnitCom) Inverse() Com { return UnitCom{} }

type unitInstance DType

func (ui *unitInstance) InputType() *DType  { return (*DType)(ui) }
func (ui *unitInstance) OutputType() *DType { return (*DType)(ui) }
func (unitInstance) Types()                 {}
func (unitInstance) Underdefined() Error    { return nil }
func (ui unitInstance) Run(input Input, output Output) {
	RunI(DType(ui), input, output)
}
