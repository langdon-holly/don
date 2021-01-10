package coms

import . "don/core"

func I(theType DType) Com { i := ICom(theType); return &i }

type ICom DType

func (ic *ICom) InputType() *DType  { return (*DType)(ic) }
func (ic *ICom) OutputType() *DType { return (*DType)(ic) }

func (ic *ICom) Types() Com {
	if DType(*ic).LTE(NullType) {
		return Null
	} else {
		return ic
	}
}

func (ic ICom) Underdefined() Error {
	return DType(ic).Underdefined().Context("in I")
}

func (ic ICom) Copy() Com { return &ic }

func (ic *ICom) Invert() Com { return ic }

func PipeUnit(outputChan chan<- Unit, inputChan <-chan Unit) {
	outputChan <- <-inputChan
}

func RunI(theType DType, input Input, output Output) {
	if !theType.NoUnit {
		go PipeUnit(output.Unit, input.Unit)
	}
	for fieldName, fieldType := range theType.Fields {
		go RunI(fieldType, input.Fields[fieldName], output.Fields[fieldName])
	}
}

func (ic ICom) Run(input Input, output Output) {
	RunI(DType(ic), input, output)
}
