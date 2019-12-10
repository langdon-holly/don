package coms

import . "don/core"

func RunI(theType DType, inputGetter InputGetter, outputGetter OutputGetter) {
	switch theType.Tag {
	case UnitTypeTag:
		inputGetter.Unit <- <-outputGetter.Unit
	case RefTypeTag:
		inputGetter.Ref <- <-outputGetter.Ref
	case StructTypeTag:
		for fieldName, fieldType := range theType.Fields {
			go RunI(fieldType, inputGetter.Struct[fieldName], outputGetter.Struct[fieldName])
		}
	}
	return

	//input := inputGetter.GetInput(theType)
	//output := outputGetter.GetOutput(theType)
	//
	//switch theType.Tag {
	//case UnitTypeTag:
	//	for {
	//		select {
	//		case <-input.Unit:
	//			output.WriteUnit()
	//		case <-quit:
	//			return
	//		}
	//	}
	//case RefTypeTag:
	//	for {
	//		select {
	//		case val := <-input.Ref:
	//			output.WriteRef(val)
	//		case <-quit:
	//			return
	//		}
	//	}
	//case StructTypeTag:
	//	for fieldName, fieldType := range theType.Fields {
	//		go RunI(fieldType, input.Struct[fieldName], output.Struct[fieldName], quit)
	//	}
	//}
}

type ICom struct{}

func (ICom) OutputType(inputType PartialType) PartialType { return inputType }

func (ICom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	RunI(inputType, inputGetter, outputGetter)
}
