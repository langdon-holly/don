package coms

import . "don/core"

type NullCom struct{}

func (NullCom) Instantiate() ComInstance { return nullInstance{} }

type nullInstance struct{}

func (nullInstance) InputType() *DType              { return NullPtr() }
func (nullInstance) OutputType() *DType             { return NullPtr() }
func (nullInstance) Types()                         {}
func (nullInstance) Underdefined() Error            { return nil }
func (nullInstance) Run(input Input, output Output) {}
