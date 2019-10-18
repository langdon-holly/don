package coms

import . "don/core"

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
	case RefTypeTag:
		a, b := inputA.Ref, inputB.Ref
		o := output.Ref
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
	case ComTypeTag:
		a, b := inputA.Com, inputB.Com
		o := output.Com
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

type MergeCom struct{}

func (MergeCom) OutputType(inputType PartialType) PartialType {
	if inputType.P {
		return MergePartialTypes(inputType.Fields["a"], inputType.Fields["b"])
	} else {
		return PartialType{}
	}
}

func (MergeCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	runMerge(inputType.Fields["a"], input.Struct["a"], input.Struct["b"], output, quit)
}
