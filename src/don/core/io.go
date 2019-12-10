package core

type Unit struct{}

type Ref struct {
	P     bool
	Input /* for P */
}

type Input struct {
	Unit   <-chan Unit
	Ref    <-chan Ref
	Struct map[string]Input
}

type Output struct {
	Unit   chan<- Unit
	Ref    chan<- Ref
	Struct map[string]Output
}

type InputGetter struct {
	Unit   chan<- chan<- Unit
	Ref    chan<- chan<- Ref
	Struct map[string]InputGetter
}

type OutputGetter struct {
	Unit   <-chan chan<- Unit
	Ref    <-chan chan<- Ref
	Struct map[string]OutputGetter
}

func MakeIO(theType DType) (inputGetter InputGetter, outputGetter OutputGetter) {
	switch theType.Tag {
	case UnitTypeTag:
		theChan := make(chan chan<- Unit, 1)
		inputGetter.Unit = theChan
		outputGetter.Unit = theChan
	case RefTypeTag:
		theChan := make(chan chan<- Ref, 1)
		inputGetter.Ref = theChan
		outputGetter.Ref = theChan
	case StructTypeTag:
		inputGetter.Struct = make(map[string]InputGetter)
		outputGetter.Struct = make(map[string]OutputGetter)
		for fieldName, fieldType := range theType.Fields {
			inputGetter.Struct[fieldName], outputGetter.Struct[fieldName] = MakeIO(fieldType)
		}
	}
	return
}

func (ig InputGetter) GetInput(theType DType) (input Input) {
	switch theType.Tag {
	case UnitTypeTag:
		theChan := make(chan Unit, 1)
		ig.Unit <- chan<- Unit(theChan)
		input.Unit = theChan
	case RefTypeTag:
		theChan := make(chan Ref, 1)
		ig.Ref <- chan<- Ref(theChan)
		input.Ref = theChan
	case StructTypeTag:
		input.Struct = make(map[string]Input)
		for fieldName, fieldType := range theType.Fields {
			input.Struct[fieldName] = ig.Struct[fieldName].GetInput(fieldType)
		}
	}
	return
}

func (og OutputGetter) GetOutput(theType DType) (output Output) {
	switch theType.Tag {
	case UnitTypeTag:
		output.Unit = <-og.Unit
	case RefTypeTag:
		output.Ref = <-og.Ref
	case StructTypeTag:
		output.Struct = make(map[string]Output)
		for fieldName, fieldType := range theType.Fields {
			output.Struct[fieldName] = og.Struct[fieldName].GetOutput(fieldType)
		}
	}
	return
}

func (o Output) WriteUnit() {
	o.Unit <- Unit{}
}

func (o Output) WriteRef(val Ref) {
	o.Ref <- val
}
