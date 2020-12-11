package coms

import "strconv"

import . "don/core"

type ParCom []Com

func (pc ParCom) Instantiate() ComInstance {
	inners := make(map[int]ComInstance, len(pc))
	ioType := MakeNFieldsType(len(pc))
	for i, subCom := range pc {
		inners[i] = subCom.Instantiate()
		ioType.Fields[strconv.Itoa(i)] = UnknownType
	}
	return &parInstance{Inners: inners, inputType: ioType, outputType: ioType}
}

func (pc ParCom) Inverse() Com {
	subInverses := make([]Com, len(pc))
	for i, subCom := range pc {
		subInverses[i] = subCom.Inverse()
	}
	return ParCom(subInverses)
}

type parInstance struct {
	Inners                map[int]ComInstance
	inputType, outputType DType
	Typesed               bool
}

func (pi *parInstance) InputType() *DType  { return &pi.inputType }
func (pi *parInstance) OutputType() *DType { return &pi.outputType }

func (pi *parInstance) Types() {
	for i, inner := range pi.Inners {
		idxStr := strconv.Itoa(i)
		newInputType := pi.inputType.Get(idxStr)
		newOutputType := pi.outputType.Get(idxStr)
		if !pi.Typesed ||
			!inner.InputType().LTE(newInputType) ||
			!inner.OutputType().LTE(newOutputType) {
			inner.InputType().Meets(newInputType)
			inner.OutputType().Meets(newOutputType)
			inner.Types()
			pi.inputType.Meets(inner.InputType().At(idxStr))
			pi.outputType.Meets(inner.OutputType().At(idxStr))
			if inner.InputType().LTE(NullType) {
				delete(pi.Inners, i)
			} else {
				pi.Inners[i] = inner
			}
		}
	}
	pi.Typesed = true
}

func (pi parInstance) Underdefined() (underdefined Error) {
	for i, inner := range pi.Inners {
		underdefined.Ors(inner.Underdefined().Context("in " + strconv.Itoa(i) + "'th computer in par"))
	}
	return
}

func (pi parInstance) Run(input Input, output Output) {
	for i, inner := range pi.Inners {
		idxStr := strconv.Itoa(i)
		go inner.Run(input.Fields[idxStr], output.Fields[idxStr])
	}
}
