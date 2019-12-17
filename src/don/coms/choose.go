package coms

import . "don/core"

type ChooseCom struct{}

func (ChooseCom) OutputType(inputType DType) DType {
	ret := MakeStructType(nil)
	if inputType.P {
		ret = MergeTypes(ret, inputType.Fields["choices"])
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
