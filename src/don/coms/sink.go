package coms

import . "don/core"

type SinkCom DType

func (com SinkCom) InputType() DType {
	return DType(com)
}

func (com SinkCom) OutputType() DType {
	return MakeStructType(make(map[string]DType, 0))
}

func RunSink(theType DType, input Input, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		i := input.Unit
		for {
			select {
			case <-i:
			case <-quit:
				return
			}
		}
	case SyntaxTypeTag:
		i := input.Syntax
		for {
			select {
			case <-i:
			case <-quit:
				return
			}
		}
	case GenComTypeTag:
		i := input.GenCom
		for {
			select {
			case <-i:
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		i := input.Struct
		for fieldName, fieldType := range theType.Fields {
			go RunSink(fieldType, i[fieldName], quit)
		}
	}
}

func (com SinkCom) Run(input Input, output Output, quit <-chan struct{}) {
	RunSink(DType(com), input, quit)
}

type GenSink struct{}

func (GenSink) OutputType(inputType PartialType) PartialType {
	return PartialType{
		P:      true,
		Tag:    StructTypeTag,
		Fields: make(map[string]PartialType, 0)}
}

func (GenSink) Com(inputType DType) Com { return SinkCom(inputType) }
