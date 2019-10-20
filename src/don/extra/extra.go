package extra

import . "don/core"

func MakeIOChans(theType DType, nInputs int) (inputs []Input, output Output) {
	inputs = make([]Input, nInputs)
	switch theType.Tag {
	case UnitTypeTag:
		output.Unit = make([]chan<- Unit, nInputs)
		for i := 0; i < nInputs; i++ {
			theChan := make(chan Unit, 1)
			inputs[i].Unit = theChan
			output.Unit[i] = theChan
		}
	case RefTypeTag:
		output.Ref = make([]chan<- Ref, nInputs)
		for i := 0; i < nInputs; i++ {
			theChan := make(chan Ref, 1)
			inputs[i].Ref = theChan
			output.Ref[i] = theChan
		}
	case StructTypeTag:
		for i := 0; i < nInputs; i++ {
			inputs[i].Struct = make(map[string]Input)
		}
		output.Struct = make(map[string]Output)
		for fieldName, fieldType := range theType.Fields {
			subInputs, subOutput := MakeIOChans(fieldType, nInputs)
			for i := 0; i < nInputs; i++ {
				inputs[i].Struct[fieldName] = subInputs[i]
			}
			output.Struct[fieldName] = subOutput
		}
	}
	return
}

func Run(com Com, inputType DType, outputIN int) (inputO Output, outputIs []Input, quit chan<- struct{}) {
	var inputIs []Input
	var outputO Output

	inputIs, inputO = MakeIOChans(inputType, 1)
	outputIs, outputO = MakeIOChans(
		HolizePartialType(com.OutputType(PartializeType(inputType))),
		outputIN)

	quitChan := make(chan struct{})
	quit = quitChan

	go com.Run(inputType, inputIs[0], outputO, quitChan)

	return
}
