package coms

import . "don/core"

type ICom struct{}

func PipeUnit(outputChan chan<- Unit, inputChan <-chan Unit) {
	for {
		<-inputChan
		outputChan <- Unit{}
	}
}

func (ICom) Types(inputType, outputType *DType) (underdefined Error) {
	inputType.Meets(*outputType)
	*outputType = *inputType
	return inputType.Underdefined().Context("in input to I")
}

func (ICom) Run(inputType, outputType DType, input Input, output Output) {
	if !inputType.NoUnit {
		go PipeUnit(output.Unit, input.Unit)
	}
	for fieldName, fieldType := range inputType.Fields {
		go ICom{}.Run(fieldType, fieldType, input.Fields[fieldName], output.Fields[fieldName])
	}
	return
}
