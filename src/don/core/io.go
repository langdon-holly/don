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

// t.Positive
func MakeIO(t DType) (input Input, output Output) {
	if !t.NoUnit {
		theChan := make(chan Unit, 1)
		input.Unit = theChan
		output.Unit = theChan
	}
	input.Fields = make(map[string]Input)
	output.Fields = make(map[string]Output)
	for fieldName, fieldType := range t.Fields {
		input.Fields[fieldName], output.Fields[fieldName] = MakeIO(fieldType)
	}
	return
}

func (o Output) WriteUnit() { o.Unit <- Unit{} }
