package coms

import "strconv"

import . "don/core"

/* Nonempty */
type PipeCom []Com

func (pc PipeCom) Instantiate() ComInstance {
	inners := make([]ComInstance, len(pc))
	toType := make(map[int]struct{}, len(pc))
	for i, subCom := range pc {
		inners[i] = subCom.Instantiate()
		toType[i] = struct{}{}
	}
	return &pipeInstance{Inners: inners, ToType: toType}
}

type pipeInstance struct {
	Inners                []ComInstance
	inputType, outputType DType
	ToType                map[int]struct{}
}

func (pi *pipeInstance) InputType() *DType  { return &pi.inputType }
func (pi *pipeInstance) OutputType() *DType { return &pi.outputType }

func (pi *pipeInstance) Types() {
	if !pi.Inners[0].InputType().LTE(pi.inputType) {
		pi.Inners[0].InputType().Meets(pi.inputType)
		pi.ToType[0] = struct{}{}
	}
	lastIdx := len(pi.Inners) - 1
	if !pi.Inners[lastIdx].OutputType().LTE(pi.outputType) {
		pi.Inners[lastIdx].OutputType().Meets(pi.outputType)
		pi.ToType[lastIdx] = struct{}{}
	}
	for i := 0; len(pi.ToType) > 0; {
		for i = range pi.ToType {
			break
		}
		delete(pi.ToType, i)

		pi.Inners[i].Types()

		if i < lastIdx && !pi.Inners[i+1].InputType().LTE(*pi.Inners[i].OutputType()) {
			pi.Inners[i+1].InputType().Meets(*pi.Inners[i].OutputType())
			pi.ToType[i+1] = struct{}{}
		}
		if i > 0 && !pi.Inners[i-1].OutputType().LTE(*pi.Inners[i].InputType()) {
			pi.Inners[i-1].OutputType().Meets(*pi.Inners[i].InputType())
			pi.ToType[i-1] = struct{}{}
		}
	}
	pi.inputType = *pi.Inners[0].InputType()
	pi.outputType = *pi.Inners[lastIdx].OutputType()
}

func (pi pipeInstance) Underdefined() (underdefined Error) {
	for i, inner := range pi.Inners {
		underdefined.Ors(inner.Underdefined().Context("in " + strconv.Itoa(i) + "'th computer in pipe"))
	}
	return
}

func (pi pipeInstance) Run(input Input, output Output) {
	currOutput := output
	for i := len(pi.Inners) - 1; i > 0; i-- {
		currInput, nextOutput := MakeIO(*pi.Inners[i].InputType())
		go pi.Inners[i].Run(currInput, currOutput)
		currOutput = nextOutput
	}
	go pi.Inners[0].Run(input, currOutput)
}
