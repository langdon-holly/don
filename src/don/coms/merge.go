package coms

import . "don/core"

type MergeCom DType

func (com MergeCom) InputType() DType {
	theType := DType(com)
	fields := make(map[string]DType, 2)
	fields["a"] = theType
	fields["b"] = theType
	return DType{StructTypeTag, fields}
}

func (com MergeCom) OutputType() DType {
	return DType(com)
}

func runMerge(theType DType, inputA, inputB interface{}, output interface{}, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		a, b := inputA.(<-chan Unit), inputB.(<-chan Unit)
		o := output.(chan<- Unit)
		for {
			select {
			case <-a:
				o <- Unit{}
			case <-b:
				o <- Unit{}
			case <-quit:
				return
			}
		}
	case SyntaxTypeTag:
		a, b := inputA.(<-chan Syntax), inputB.(<-chan Syntax)
		o := output.(chan<- Syntax)
		for {
			select {
			case v := <-a:
				o <- v
			case v := <-b:
				o <- v
			case <-quit:
				return
			}
		}
	case GenComTypeTag:
		a, b := inputA.(<-chan GenCom), inputB.(<-chan GenCom)
		o := output.(chan<- GenCom)
		for {
			select {
			case v := <-a:
				o <- v
			case v := <-b:
				o <- v
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		a, b := inputA.(Struct), inputB.(Struct)
		o := output.(Struct)
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			go runMerge(fieldType, a[fieldName], b[fieldName], o[fieldName], quit)
		}
	}
}

func (com MergeCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	inputStruct := input.(Struct)
	runMerge(DType(com), inputStruct["a"], inputStruct["b"], output, quit)
}

func GenMerge(inputType DType) Com { return MergeCom(inputType) }
