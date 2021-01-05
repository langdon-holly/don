package coms

import . "don/core"

type ICom DType

func (ic ICom) Instantiate() ComInstance { ii := iInstance(ic); return &ii }
func (ic ICom) Inverse() Com             { return ic }

type iInstance DType

func (ii *iInstance) InputType() *DType  { return (*DType)(ii) }
func (ii *iInstance) OutputType() *DType { return (*DType)(ii) }

func (ii iInstance) Types() {}

func (ii iInstance) Underdefined() Error {
	return DType(ii).Underdefined().Context("in I")
}

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

func (ii iInstance) Run(input Input, output Output) {
	RunI(DType(ii), input, output)
}
