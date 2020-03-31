package coms

import . "don/core"

type ChooseCom struct{}

func (ChooseCom) OutputType(inputType DType) (outputType DType, impossible bool) {
	outputType = MakeStructType(nil)
	if inputType.Tag != UnknownTypeTag {
		if inputType.Tag != StructTypeTag {
			impossible = true
			return
		}

		readyType := inputType.Fields["ready"]
		if readyType.Tag != UnknownTypeTag && readyType.Tag != UnitTypeTag {
			impossible = true
			return
		}

		outputType, impossible = MergeTypes(outputType, inputType.Fields["choices"])
	}
	return
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
