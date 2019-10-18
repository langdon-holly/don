package extra

import . "don/core"

func MakeIOChans(theType DType) (input Input, output Output) {
	switch theType.Tag {
	case UnitTypeTag:
		theChan := make(chan Unit, 1)
		input.Unit = theChan
		output.Unit = theChan
	case SyntaxTypeTag:
		theChan := make(chan Syntax, 1)
		input.Syntax = theChan
		output.Syntax = theChan
	case GenComTypeTag:
		theChan := make(chan GenCom, 1)
		input.GenCom = theChan
		output.GenCom = theChan
	case StructTypeTag:
		input.Struct = make(map[string]Input)
		output.Struct = make(map[string]Output)
		for fieldName, fieldType := range theType.Fields {
			input.Struct[fieldName], output.Struct[fieldName] = MakeIOChans(fieldType)
		}
	}
	return
}

func Run(com Com) (inputO Output, outputI Input, quit chan<- struct{}) {
	var inputI Input
	var outputO Output

	inputI, inputO = MakeIOChans(com.InputType())
	outputI, outputO = MakeIOChans(com.OutputType())

	quitChan := make(chan struct{})
	quit = quitChan

	go com.Run(inputI, outputO, quitChan)

	return
}
