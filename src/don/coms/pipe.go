package coms

import "strconv"

import . "don/core"

/* Nonempty */
type PipeCom []Com

func (pc PipeCom) Instantiate() ComInstance {
	subComIs := make([]ComInstance, len(pc))
	for i, subCom := range pc {
		subComIs[i] = subCom.Instantiate()
	}
	return pipeInstance(subComIs)
}

type pipeInstance []ComInstance

func (pi pipeInstance) InputType() *DType { return pi[0].InputType() }
func (pi pipeInstance) OutputType() *DType {
	return pi[len(pi)-1].OutputType()
}

func (pi pipeInstance) Types() (underdefined Error) {
	subUnderdefineds := make([]Error, len(pi))
	for i := 0; i < len(pi); i++ { // Slightly inefficient?
		subUnderdefineds[i] = pi[i].Types()
		if i < len(pi)-1 {
			pi[i+1].InputType().Meets(*pi[i].OutputType())
		}
		if i > 0 && !pi[i-1].OutputType().LTE(*pi[i].InputType()) {
			pi[i-1].OutputType().Meets(*pi[i].InputType())
			i -= 2
		}
	}
	for i, subUnderdefined := range subUnderdefineds {
		underdefined.Ors(subUnderdefined.Context("in " + strconv.Itoa(i) + "'th computer in pipe"))
	}
	return
}

func (pi pipeInstance) Run(input Input, output Output) {
	currOutput := output
	for i := len(pi) - 1; i > 0; i-- {
		currInput, nextOutput := MakeIO(*pi[i].InputType())
		go pi[i].Run(currInput, currOutput)
		currOutput = nextOutput
	}
	go pi[0].Run(input, currOutput)
}
