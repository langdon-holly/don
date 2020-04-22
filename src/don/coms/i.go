package coms

import . "don/core"

func RunI(theType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		inputChan := inputGetter.GetInput(UnitType).Unit
		outputChan := outputGetter.GetOutput(UnitType).Unit
		go PipeUnit(outputChan, inputChan, quit)
	case StructTypeTag:
		for fieldName, fieldType := range theType.Fields {
			go RunI(fieldType, inputGetter.Struct[fieldName], outputGetter.Struct[fieldName], quit)
		}
	}
	return
}

type ICom struct{}

func PipeUnit(outputChan chan<- Unit, inputChan <-chan Unit, quit <-chan struct{}) {
	for {
		select {
		case <-inputChan:
			outputChan <- Unit{}
		case <-quit:
			return
		}
	}
}

func (ICom) OutputType(inputType DType) (outputType DType, impossible bool) { return inputType, false }

func (ICom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	RunI(inputType, inputGetter, outputGetter, quit)
}
