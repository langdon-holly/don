package coms

import . "don/core"

type SinkCom DType

func (com SinkCom) InputType() DType {
	return DType(com)
}

func (com SinkCom) OutputType() DType {
	return MakeStructType(make(map[string]DType, 0))
}

func RunSink(theType DType, input interface{}, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		i := input.(<-chan Unit)
		for {
			select {
			case <-i:
			case <-quit:
				return
			}
		}
	case SyntaxTypeTag:
		i := input.(<-chan Syntax)
		for {
			select {
			case <-i:
			case <-quit:
				return
			}
		}
	case GenComTypeTag:
		i := input.(<-chan GenCom)
		for {
			select {
			case <-i:
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		i := input.(Struct)
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			go RunSink(fieldType, i[fieldName], quit)
		}
	}
}

func (com SinkCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	RunSink(DType(com), input, quit)
}

func GenSink(inputType DType) Com { return SinkCom(inputType) }
