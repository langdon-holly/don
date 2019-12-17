package extra

import . "don/core"

//func MakeIOChans(theType DType) (input Input, output Output) {
//	switch theType.Tag {
//	case UnitTypeTag:
//		output.Unit = make(chan<- Unit, 1)
//		theChan := make(chan Unit, 1)
//		input.Unit = theChan
//		output.Unit = theChan
//	case RefTypeTag:
//		output.Ref = make(chan<- Ref, 1)
//		theChan := make(chan Ref, 1)
//		input.Ref = theChan
//		output.Ref = theChan
//	case StructTypeTag:
//		input.Struct = make(map[string]Input)
//		output.Struct = make(map[string]Output)
//		for fieldName, fieldType := range theType.Fields {
//			input.Struct[fieldName], output.Struct[fieldName] = MakeIOChans(fieldType)
//		}
//	}
//	return
//}

//func MakeUnitChan() (input Input, output Output) {
//	theChan := make(chan Unit, 1)
//	input.Unit = theChan
//	output.Unit = theChan
//	return
//}
//
//func MakeRefChan() (input Input, output Output) {
//	theChan := make(chan Ref, 1)
//	input.Ref = theChan
//	output.Ref = theChan
//	return
//}

func Run(com Com, inputType DType) (inputO Output, outputI Input, quit chan<- struct{}) {
	outputType := com.OutputType(inputType)

	inputIGetter, inputOGetter := MakeIO(inputType)
	outputIGetter, outputOGetter := MakeIO(outputType)

	quitChan := make(chan struct{})
	quit = quitChan

	go com.Run(inputType, inputIGetter, outputOGetter, quitChan)

	outputI = outputIGetter.GetInput(outputType)
	inputO = inputOGetter.GetOutput(inputType)

	return
}
