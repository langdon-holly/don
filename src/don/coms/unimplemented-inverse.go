package coms

import . "don/core"

type UnimplementedInverseCom struct{ Inner Com }

func (uic UnimplementedInverseCom) Instantiate() ComInstance {
	return unimplementedInverseInstance{Inner: uic.Inner.Instantiate()}
}

func (uic UnimplementedInverseCom) Inverse() Com { return uic.Inner }

type unimplementedInverseInstance struct{ Inner ComInstance }

func (uii unimplementedInverseInstance) InputType() *DType {
	return uii.Inner.OutputType()
}
func (uii unimplementedInverseInstance) OutputType() *DType {
	return uii.Inner.InputType()
}

func (uii unimplementedInverseInstance) Types() { uii.Inner.Types() }

func (uii unimplementedInverseInstance) Underdefined() Error {
	return uii.Inner.Underdefined().Context("in unimplemented inverse")
}

func (uii unimplementedInverseInstance) Run(Input, Output) {
	if !uii.Inner.InputType().LTE(NullType) {
		panic("Unimplemented")
	}
}
