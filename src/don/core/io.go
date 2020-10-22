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

func (o Output) WriteUnit() {
	if o.Unit != nil {
		o.Unit <- Unit{}
	}
}
