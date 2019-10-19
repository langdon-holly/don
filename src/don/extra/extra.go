package extra

import . "don/core"

func MakeIOChans(theType DType) (input Input, output Output) {
	switch theType.Tag {
	case UnitTypeTag:
		theChan := make(chan Unit, 1)
		input.Unit = theChan
		output.Unit = theChan
	case ComTypeTag:
		theChan := make(chan Com, 1)
		input.Com = theChan
		output.Com = theChan
	case StructTypeTag:
		input.Struct = make(map[string]Input)
		output.Struct = make(map[string]Output)
		for fieldName, fieldType := range theType.Fields {
			input.Struct[fieldName], output.Struct[fieldName] = MakeIOChans(fieldType)
		}
	}
	return
}

func Run(com Com, inputType DType) (inputO Output, outputI Input, quit chan<- struct{}) {
	var inputI Input
	var outputO Output

	inputI, inputO = MakeIOChans(inputType)
	outputI, outputO = MakeIOChans(HolizePartialType(com.OutputType(PartializeType(inputType))))

	quitChan := make(chan struct{})
	quit = quitChan

	go com.Run(inputType, inputI, outputO, quitChan)

	return
}
