package coms

import "strconv"

import . "don/core"

/* Nonempty */
type PipeCom []Com

func (pc PipeCom) pipeComTypes(inputType, outputType DType) (typeAts []DType, underdefined Error) {
	typeAts = make([]DType, len(pc)+1)
	typeAts[0] = inputType
	typeAts[len(pc)] = outputType
	subUnderdefineds := make([]Error, len(pc))
	for i := 0; i < len(pc); i++ { // Slightly inefficient?
		inputTypeBefore := typeAts[i]
		subUnderdefineds[i] = pc[i].Types(&typeAts[i], &typeAts[i+1])
		if !inputTypeBefore.LTE(typeAts[i]) && i > 0 {
			i -= 2
		}
	}
	for i, subUnderdefined := range subUnderdefineds {
		underdefined.Ors(subUnderdefined.Context("in " + strconv.Itoa(i) + "'th computer in pipe"))
	}
	return
}

func (pc PipeCom) Types(inputType, outputType *DType) (underdefined Error) {
	var typeAts []DType
	typeAts, underdefined = pc.pipeComTypes(*inputType, *outputType)
	*inputType = typeAts[0]
	*outputType = typeAts[len(pc)]
	return
}

func (pc PipeCom) Run(inputType, outputType DType, input Input, output Output) {
	typeAts, _ := pc.pipeComTypes(inputType, outputType)

	currOutput := output
	for i := len(pc) - 1; i > 0; i-- {
		currInput, nextOutput := MakeIO(typeAts[i])
		go pc[i].Run(typeAts[i], typeAts[i+1], currInput, currOutput)
		currOutput = nextOutput
	}
	go pc[0].Run(typeAts[0], typeAts[1], input, currOutput)
}
