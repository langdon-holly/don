package coms

import . "don/core"

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
	case RefTypeTag:
		i := input.Ref
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

type SinkCom struct{}

func (SinkCom) OutputType(inputType PartialType) PartialType {
	return PartialType{
		P:      true,
		Tag:    StructTypeTag,
		Fields: make(map[string]PartialType, 0)}
}

func (SinkCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	RunSink(inputType, input, quit)
}
