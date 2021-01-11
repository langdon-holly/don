package coms

import . "don/core"

var Null = NullCom{}

type NullCom struct{}

func (NullCom) InputType() DType           { return NullType }
func (NullCom) OutputType() DType          { return NullType }
func (NullCom) MeetTypes(DType, DType) Com { return NullCom{} }
func (NullCom) Underdefined() Error        { return nil }
func (NullCom) Copy() Com                  { return NullCom{} }
func (NullCom) Invert() Com                { return NullCom{} }
func (NullCom) Run(Input, Output)          {}
