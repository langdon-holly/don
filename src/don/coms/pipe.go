package coms

import "strconv"

import . "don/core"

/* Nonempty */
type PipeCom []Com

func (pc PipeCom) pipeComTypes(inputType, outputType DType) (typeAts []DType, bad []string, done bool) {
	typeAts = make([]DType, len(pc)+1)
	typeAts[0] = inputType
	typeAts[len(pc)] = outputType
	subDones := make([]bool, len(pc))
	for i := 0; i < len(pc); i++ { // Slightly inefficient?
		inputTypeBefore := typeAts[i]
		bad, subDones[i] = pc[i].Types(&typeAts[i], &typeAts[i+1])
		if bad != nil {
			bad = append(bad, "in pipe subcomputer "+strconv.FormatInt(int64(i), 10))
			return
		}
		if !inputTypeBefore.Equal(typeAts[i]) && i > 0 {
			i -= 2
		}
	}

	done = true
	for _, subDone := range subDones {
		done = done && subDone
	}

	return
}

func (pc PipeCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	var typeAts []DType
	typeAts, bad, done = pc.pipeComTypes(*inputType, *outputType)
	*inputType = typeAts[0]
	*outputType = typeAts[len(pc)]
	return
}

func (pc PipeCom) Run(inputType, outputType DType, input Input, output Output) {
	typeAts, _, _ := pc.pipeComTypes(inputType, outputType)

	currOutput := output
	for i := len(pc) - 1; i > 0; i-- {
		currInput, nextOutput := MakeIO(typeAts[i])
		go pc[i].Run(typeAts[i], typeAts[i+1], currInput, currOutput)
		currOutput = nextOutput
	}
	go pc[0].Run(typeAts[0], typeAts[1], input, currOutput)
}
