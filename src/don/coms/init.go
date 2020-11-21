package coms

import . "don/core"

type InitCom struct{}

func (InitCom) Instantiate() ComInstance {
	ii := initInstance(UnitType)
	return &ii
}

type initInstance DType

func (ii initInstance) InputType() *DType   { return NullPtr() }
func (ii *initInstance) OutputType() *DType { return (*DType)(ii) }

// Violates multiplicative annihilation!!
func (ii initInstance) Types() {}

func (ii initInstance) Underdefined() Error { return nil }

func (ii initInstance) Run(input Input, output Output) {
	if !DType(ii).NoUnit {
		output.WriteUnit()
	}
}
