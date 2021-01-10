package coms

import "strconv"

import . "don/core"

/* Nonempty inners */
func Pipe(inners []Com) Com {
	toType := make(map[int]struct{}, len(inners))
	for i := range inners {
		toType[i] = struct{}{}
	}
	for i := 0; len(toType) > 0; {
		for i = range toType {
			break
		}
		delete(toType, i)

		inners[i] = inners[i].Types()

		if i < len(inners)-1 && !inners[i+1].InputType().LTE(*inners[i].OutputType()) {
			inners[i+1].InputType().Meets(*inners[i].OutputType())
			toType[i+1] = struct{}{}
		}
		if i > 0 && !inners[i-1].OutputType().LTE(*inners[i].InputType()) {
			inners[i-1].OutputType().Meets(*inners[i].InputType())
			toType[i-1] = struct{}{}
		}
	}
	hasNull := false
	firstNullInputType := 0
	lastNullOutputType := 0
	for i, inner := range inners {
		if inner.InputType().LTE(NullType) {
			firstNullInputType = i
			for j := len(inners) - 1; j >= i; j-- {
				if inners[j].OutputType().LTE(NullType) {
					lastNullOutputType = j
					hasNull = true
					break
				}
			}
			break
		}
	}
	if hasNull {
		inners = append(
			append(inners[:firstNullInputType], Null),
			inners[lastNullOutputType+1:]...)
	}
	if len(inners) == 1 {
		return inners[0]
	} else {
		inputType := *inners[0].InputType()
		outputType := *inners[len(inners)-1].OutputType()
		return &PipeCom{Inners: inners, inputType: inputType, outputType: outputType}
	}
}

/* Nonempty Inners */
type PipeCom struct {
	Inners                []Com
	inputType, outputType DType
}

func (pc *PipeCom) InputType() *DType  { return &pc.inputType }
func (pc *PipeCom) OutputType() *DType { return &pc.outputType }

func (pc *PipeCom) Types() Com {
	toType := make(map[int]struct{})
	if !pc.Inners[0].InputType().LTE(pc.inputType) {
		pc.Inners[0].InputType().Meets(pc.inputType)
		toType[0] = struct{}{}
	}
	lastIdx := len(pc.Inners) - 1
	if !pc.Inners[lastIdx].OutputType().LTE(pc.outputType) {
		pc.Inners[lastIdx].OutputType().Meets(pc.outputType)
		toType[lastIdx] = struct{}{}
	}
	for i := 0; len(toType) > 0; {
		for i = range toType {
			break
		}
		delete(toType, i)

		pc.Inners[i] = pc.Inners[i].Types()

		if i < lastIdx && !pc.Inners[i+1].InputType().LTE(*pc.Inners[i].OutputType()) {
			pc.Inners[i+1].InputType().Meets(*pc.Inners[i].OutputType())
			toType[i+1] = struct{}{}
		}
		if i > 0 && !pc.Inners[i-1].OutputType().LTE(*pc.Inners[i].InputType()) {
			pc.Inners[i-1].OutputType().Meets(*pc.Inners[i].InputType())
			toType[i-1] = struct{}{}
		}
	}
	hasNull := false
	firstNullInputType := 0
	lastNullOutputType := 0
	for i, inner := range pc.Inners {
		if inner.InputType().LTE(NullType) {
			firstNullInputType = i
			for j := len(pc.Inners) - 1; j >= i; j-- {
				if pc.Inners[j].OutputType().LTE(NullType) {
					lastNullOutputType = j
					hasNull = true
					break
				}
			}
			break
		}
	}
	if hasNull {
		pc.Inners = append(
			append(pc.Inners[:firstNullInputType], Null),
			pc.Inners[lastNullOutputType+1:]...)
	}
	if len(pc.Inners) == 1 {
		return pc.Inners[0]
	} else {
		pc.inputType = *pc.Inners[0].InputType()
		pc.outputType = *pc.Inners[len(pc.Inners)-1].OutputType()
		return pc
	}
}

func (pc PipeCom) Underdefined() (underdefined Error) {
	for i, inner := range pc.Inners {
		underdefined.Ors(inner.Underdefined().Context("in " + strconv.Itoa(i) + "'th computer in pipe"))
	}
	return
}

func (pc PipeCom) Copy() Com {
	inners := make([]Com, len(pc.Inners))
	for i, inner := range pc.Inners {
		inners[i] = inner.Copy()
	}
	pc.Inners = inners
	return &pc
}

func (pc *PipeCom) Invert() Com {
	innerInverses := make([]Com, len(pc.Inners))
	for i, inner := range pc.Inners {
		innerInverses[len(pc.Inners)-1-i] = inner.Invert()
	}
	pc.Inners = innerInverses

	pc.inputType, pc.outputType = pc.outputType, pc.inputType
	return pc
}

func (pc PipeCom) Run(input Input, output Output) {
	currOutput := output
	for i := len(pc.Inners) - 1; i > 0; i-- {
		currInput, nextOutput := MakeIO(*pc.Inners[i].InputType())
		go pc.Inners[i].Run(currInput, currOutput)
		currOutput = nextOutput
	}
	go pc.Inners[0].Run(input, currOutput)
}
