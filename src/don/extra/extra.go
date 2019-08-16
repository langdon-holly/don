package extra

import . "don/core"

func MakeIOChans(theType DType) (input, output interface{}) {
	switch theType.Tag {
	case UnitTypeTag:
		theChan := make(chan Unit, 1)
		input = (<-chan Unit)(theChan)
		output = chan<- Unit(theChan)
	case SyntaxTypeTag:
		theChan := make(chan Syntax, 1)
		input = (<-chan Syntax)(theChan)
		output = chan<- Syntax(theChan)
	case GenComTypeTag:
		theChan := make(chan GenCom, 1)
		input = (<-chan GenCom)(theChan)
		output = chan<- GenCom(theChan)
	case StructTypeTag:
		inputMap := make(Struct)
		outputMap := make(Struct)
		input = inputMap
		output = outputMap
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			inputMap[fieldName], outputMap[fieldName] = MakeIOChans(fieldType)
		}
	}
	return
}

func Run(com Com) (inputO, outputI interface{}, quit chan<- struct{}) {
	var inputI, outputO interface{}

	inputI, inputO = MakeIOChans(com.InputType())
	outputI, outputO = MakeIOChans(com.OutputType())

	quitChan := make(chan struct{})
	quit = quitChan

	go com.Run(inputI, outputO, quitChan)

	return
}
