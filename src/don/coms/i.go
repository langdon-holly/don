package coms

import . "don/core"

func RunI(theType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		inputChan := inputGetter.GetInput(UnitType).Unit
		outputChan := outputGetter.GetOutput(UnitType).Unit
		go PipeUnit(outputChan, inputChan, quit)
	case RefTypeTag:
		inputGetter.Ref <- <-outputGetter.Ref
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

func PipeRef(outputChan chan<- Ref, inputChan <-chan Ref, quit <-chan struct{}) {
	for {
		select {
		case val := <-inputChan:
			outputChan <- val
		case <-quit:
			return
		}
	}
}

func (ICom) OutputType(inputType DType) DType { return inputType }

func (ICom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	RunI(inputType, inputGetter, outputGetter, quit)
}
