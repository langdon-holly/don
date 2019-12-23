package coms

import . "don/core"

type ChooseCom struct{}

func (ChooseCom) OutputType(inputType DType) DType {
	ret := MakeStructType(nil)
	switch inputType.Lvl {
	case UnknownTypeLvl:
	case NormalTypeLvl:
		if inputType.Tag != StructTypeTag {
			return ImpossibleType
		}

		readyType := inputType.Fields["ready"]
		switch readyType.Lvl {
		case UnknownTypeLvl:
		case NormalTypeLvl:
			if readyType.Tag != UnitTypeTag {
				return ImpossibleType
			}
		case ImpossibleTypeLvl:
			return ImpossibleType
		}

		ret = MergeTypes(ret, inputType.Fields["choices"])
	case ImpossibleTypeLvl:
		return ImpossibleType
	}
	return ret
}

func listen(chosens chan<- string, fieldName string, choice <-chan Unit, quit <-chan struct{}) {
	for {
		select {
		case <-choice:
			chosens <- fieldName
		case <-quit:
			return
		}
	}
}

func (ChooseCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	choicesIn := input.Struct["choices"].Struct
	ready := input.Struct["ready"].Unit
	choicesOut := output.Struct

	chosens := make(chan string)

	for fieldName, _ := range inputType.Fields["choices"].Fields {
		go listen(chosens, fieldName, choicesIn[fieldName].Unit, quit)
	}

	for {
		select {
		case <-ready:
		case <-quit:
			return
		}
		select {
		case chosen := <-chosens:
			choicesOut[chosen].WriteUnit()
		case <-quit:
			return
		}
	}
}
