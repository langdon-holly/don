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

func runMerge(theType DType, inputA, inputB Input, output Output, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		a, b := inputA.Unit, inputB.Unit
		o := output.Unit
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
		a, b := inputA.Syntax, inputB.Syntax
		o := output.Syntax
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
		a, b := inputA.GenCom, inputB.GenCom
		o := output.GenCom
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
		a, b := inputA.Struct, inputB.Struct
		o := output.Struct
		for fieldName, fieldType := range theType.Fields {
			go runMerge(fieldType, a[fieldName], b[fieldName], o[fieldName], quit)
		}
	}
}

func (com MergeCom) Run(input Input, output Output, quit <-chan struct{}) {
	runMerge(DType(com), input.Struct["a"], input.Struct["b"], output, quit)
}

type GenMerge struct{}

func (GenMerge) OutputType(inputType PartialType) PartialType {
	if inputType.P {
		return MergePartialTypes(inputType.Fields["a"], inputType.Fields["b"])
	} else {
		return PartialType{}
	}
}

func (GenMerge) Com(inputType DType) Com { return MergeCom(inputType.Fields["a"]) }
