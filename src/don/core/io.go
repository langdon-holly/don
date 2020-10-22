package core

type Unit struct{}

type Input struct {
	Unit   <-chan Unit
	Fields map[string]Input
}

type Output struct {
	Unit   chan<- Unit
	Fields map[string]Output
}

func MakeIO(theType DType) (input Input, output Output) {
	if theType.Tag == UnitTypeTag {
		theChan := make(chan Unit, 1)
		input.Unit = theChan
		output.Unit = theChan
	} else { /* theType.Tag == StructTypeTag */
		input.Fields = make(map[string]Input)
		output.Fields = make(map[string]Output)
		for fieldName, fieldType := range theType.Fields {
			input.Fields[fieldName], output.Fields[fieldName] = MakeIO(fieldType)
		}
	}
	return
}

//type InputGetter struct {
//	Unit   chan<- chan<- Unit
//	Fields map[string]InputGetter
//}
//
//type OutputGetter struct {
//	Unit   <-chan chan<- Unit
//	Fields map[string]OutputGetter
//}
//
//func MakeIO(theType DType) (inputGetter InputGetter, outputGetter OutputGetter) {
//	if theType.Tag == UnitTypeTag {
//		theChan := make(chan chan<- Unit, 1)
//		inputGetter.Unit = theChan
//		outputGetter.Unit = theChan
//	} else { /* theType.Tag == StructTypeTag */
//
//		inputGetter.Fields = make(map[string]InputGetter)
//		outputGetter.Fields = make(map[string]OutputGetter)
//		for fieldName, fieldType := range theType.Fields {
//			inputGetter.Fields[fieldName], outputGetter.Fields[fieldName] = MakeIO(fieldType)
//		}
//	}
//
//	return
//}
//
//func (ig InputGetter) GetInput() (input Input) {
//	if ig.Unit != nil {
//		theChan := make(chan Unit, 1)
//		ig.Unit <- chan<- Unit(theChan)
//		input.Unit = theChan
//	}
//
//	input.Fields = make(map[string]Input, len(ig.Fields))
//	for fieldName, subIg := range ig.Fields {
//		input.Fields[fieldName] = subIg.GetInput()
//	}
//
//	return
//}
//
//func (ig InputGetter) SendOutput(output Output) {
//	if ig.Unit != nil {
//		ig.Unit <- output.Unit
//	}
//
//	for fieldName, subIg := range ig.Fields {
//		subIg.SendOutput(output.Fields[fieldName])
//	}
//	return
//}
//
//func (og OutputGetter) GetOutput() (output Output) {
//	if og.Unit != nil {
//		output.Unit = <-og.Unit
//	}
//
//	output.Fields = make(map[string]Output, len(og.Fields))
//	for fieldName, subOg := range og.Fields {
//		output.Fields[fieldName] = subOg.GetOutput()
//	}
//
//	return
//}

func (o Output) WriteUnit() {
	if o.Unit != nil {
		o.Unit <- Unit{}
	}
}
