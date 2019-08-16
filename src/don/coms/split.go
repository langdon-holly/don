package coms

import . "don/core"

type SplitCom DType

func (com SplitCom) InputType() DType {
	return DType(com)
}

func (com SplitCom) OutputType() DType {
	theType := DType(com)
	fields := make(map[string]DType, 2)
	fields["a"] = theType
	fields["b"] = theType
	return DType{StructTypeTag, fields}
}

func runSplit(theType DType, input interface{}, outputA, outputB interface{}, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		i := input.(<-chan Unit)
		a, b := outputA.(chan<- Unit), outputB.(chan<- Unit)
		for {
			select {
			case <-i:
				a <- Unit{}
				b <- Unit{}
			case <-quit:
				return
			}
		}
	case SyntaxTypeTag:
		i := input.(<-chan Syntax)
		a, b := outputA.(chan<- Syntax), outputB.(chan<- Syntax)
		for {
			select {
			case v := <-i:
				a <- v
				b <- v
			case <-quit:
				return
			}
		}
	case GenComTypeTag:
		i := input.(<-chan GenCom)
		a, b := outputA.(chan<- GenCom), outputB.(chan<- GenCom)
		for {
			select {
			case v := <-i:
				a <- v
				b <- v
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		i := input.(Struct)
		a, b := outputA.(Struct), outputB.(Struct)
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			go runSplit(fieldType, i[fieldName], a[fieldName], b[fieldName], quit)
		}
	}
}

func (com SplitCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	outputStruct := output.(Struct)
	runSplit(DType(com), input, outputStruct["a"], outputStruct["b"], quit)
}

func GenSplit(inputType DType) Com { return SplitCom(inputType) }
