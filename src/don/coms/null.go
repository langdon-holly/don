package coms

import . "don/core"

var Null = NullCom{}

type NullCom struct{}

func (NullCom) InputType() *DType              { return NullPtr() }
func (NullCom) OutputType() *DType             { return NullPtr() }
func (NullCom) Types() Com                     { return NullCom{} }
func (NullCom) Underdefined() Error            { return nil }
func (NullCom) Copy() Com                      { return NullCom{} }
func (NullCom) Invert() Com                    { return NullCom{} }
func (NullCom) Run(input Input, output Output) {}
