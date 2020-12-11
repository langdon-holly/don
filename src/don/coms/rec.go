package coms

import . "don/core"

type RecCom struct{ Inner Com }

var recComInOutType = MakeNFieldsType(2)

func init() {
	recComInOutType.Fields["rec"] = UnknownType
	recComInOutType.Fields["out"] = UnknownType
}

func (rc RecCom) Instantiate() ComInstance {
	ri := recInstance{
		Gather:  GatherCom{}.Instantiate(),
		Scatter: ScatterCom{}.Instantiate(),
		Inner:   rc.Inner.Instantiate(),
		ToType:  make(map[string]struct{}, 3)}
	ri.Gather.InputType().Meets(recComInOutType)
	ri.Scatter.OutputType().Meets(recComInOutType)
	ri.ToType["gather"] = struct{}{}
	ri.ToType["scatter"] = struct{}{}
	ri.ToType["inner"] = struct{}{}
	return &ri
}

func (rc RecCom) Inverse() Com { return RecCom{Inner: rc.Inner.Inverse()} }

type recInstance struct {
	Gather, Scatter, Inner ComInstance
	inputType, outputType  DType
	ToType                 map[string]struct{}
}

func (ri *recInstance) InputType() *DType  { return &ri.inputType }
func (ri *recInstance) OutputType() *DType { return &ri.outputType }

func (ri *recInstance) Types() {
	if !ri.Gather.InputType().LTE(ri.inputType.At("out")) {
		ri.Gather.InputType().Meets(ri.inputType.At("out"))
		ri.ToType["gather"] = struct{}{}
	}
	if !ri.Scatter.OutputType().LTE(ri.outputType.At("out")) {
		ri.Scatter.OutputType().Meets(ri.outputType.At("out"))
		ri.ToType["scatter"] = struct{}{}
	}

	for typeNext := ""; len(ri.ToType) > 0; {
		for typeNext = range ri.ToType {
			break
		}
		switch delete(ri.ToType, typeNext); typeNext {
		case "gather":
			recTypeBefore := (*ri.Gather.InputType()).Fields["rec"]
			innerInputTypeBefore := *ri.Gather.OutputType()

			ri.Gather.Types()
			(*ri.Scatter.OutputType()).Fields["rec"] = (*ri.Gather.InputType()).Fields["rec"]

			if !recTypeBefore.LTE((*ri.Gather.InputType()).Fields["rec"]) {
				ri.Scatter.OutputType().Meets((*ri.Gather.InputType()).Fields["rec"].At("rec"))
				ri.ToType["scatter"] = struct{}{}
			}
			if !innerInputTypeBefore.LTE(*ri.Gather.OutputType()) {
				ri.Inner.InputType().Meets(*ri.Gather.OutputType())
				ri.ToType["inner"] = struct{}{}
			}
		case "scatter":
			recTypeBefore := (*ri.Scatter.OutputType()).Fields["rec"]
			innerOutputTypeBefore := *ri.Scatter.InputType()

			ri.Scatter.Types()
			(*ri.Gather.InputType()).Fields["rec"] = (*ri.Scatter.OutputType()).Fields["rec"]

			if !recTypeBefore.LTE((*ri.Scatter.OutputType()).Fields["rec"]) {
				ri.Gather.InputType().Meets((*ri.Scatter.OutputType()).Fields["rec"].At("rec"))
				ri.ToType["gather"] = struct{}{}
			}
			if !innerOutputTypeBefore.LTE(*ri.Scatter.InputType()) {
				ri.Inner.OutputType().Meets(*ri.Scatter.InputType())
				ri.ToType["inner"] = struct{}{}
			}
		case "inner":
			innerInputTypeBefore := *ri.Inner.InputType()
			innerOutputTypeBefore := *ri.Inner.OutputType()

			ri.Inner.Types()

			if !innerInputTypeBefore.LTE(*ri.Inner.InputType()) {
				ri.Gather.OutputType().Meets(*ri.Inner.InputType())
				ri.ToType["gather"] = struct{}{}
			}
			if !innerOutputTypeBefore.LTE(*ri.Inner.OutputType()) {
				ri.Scatter.InputType().Meets(*ri.Inner.OutputType())
				ri.ToType["scatter"] = struct{}{}
			}
		}
	}

	ri.inputType = (*ri.Gather.InputType()).Fields["out"]
	ri.outputType = (*ri.Scatter.OutputType()).Fields["out"]

	if ri.inputType.LTE(NullType) || ri.outputType.LTE(NullType) {
		ri.inputType = NullType
		ri.outputType = NullType
		ri.Gather.InputType().Meets(NullType)
		ri.Scatter.InputType().Meets(NullType)
		ri.Inner.InputType().Meets(NullType)
		ri.Gather.Types()
		ri.Scatter.Types()
		ri.Inner.Types()
	}
}

func (ri recInstance) Underdefined() (underdefined Error) {
	underdefined.Ors(
		ri.Gather.Underdefined().Context("in rec gather")).Ors(
		ri.Scatter.Underdefined().Context("in rec scatter")).Ors(
		ri.Inner.Underdefined().Context("in rec inner"))
	return
}

func (ri recInstance) Run(input Input, output Output) {
	innerInput, gatherOutput := MakeIO(*ri.Gather.OutputType())
	scatterInput, innerOutput := MakeIO(*ri.Scatter.InputType())
	recInput, recOutput := MakeIO((*ri.Gather.InputType()).Fields["rec"])
	scatterOutput := Output{Fields: make(map[string]Output, 2)}
	scatterOutput.Fields["rec"] = recOutput
	scatterOutput.Fields["out"] = output
	gatherInput := Input{Fields: make(map[string]Input, 2)}
	gatherInput.Fields["out"] = input
	gatherInput.Fields["rec"] = recInput

	go ri.Inner.Run(innerInput, innerOutput)
	go ri.Gather.Run(gatherInput, gatherOutput)
	go ri.Scatter.Run(scatterInput, scatterOutput)
}
