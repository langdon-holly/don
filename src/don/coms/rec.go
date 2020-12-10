package coms

import . "don/core"

type RecCom struct{ Inner Com }

var recComInOutType = MakeNStructType(2)

func init() {
	recComInOutType.Fields["rec"] = UnknownType
	recComInOutType.Fields["out"] = UnknownType
}

func (rc RecCom) Instantiate() ComInstance {
	ri := recInstance{
		Merge:  MergeCom{}.Instantiate(),
		Split:  SplitCom{}.Instantiate(),
		Inner:  rc.Inner.Instantiate(),
		ToType: make(map[string]struct{}, 3)}
	ri.Merge.InputType().Meets(recComInOutType)
	ri.Split.OutputType().Meets(recComInOutType)
	ri.ToType["merge"] = struct{}{}
	ri.ToType["split"] = struct{}{}
	ri.ToType["inner"] = struct{}{}
	return &ri
}

func (rc RecCom) Inverse() Com { panic("Unimplemented") }

type recInstance struct {
	Merge, Split, Inner   ComInstance
	inputType, outputType DType
	ToType                map[string]struct{}
}

func (ri *recInstance) InputType() *DType  { return &ri.inputType }
func (ri *recInstance) OutputType() *DType { return &ri.outputType }

func (ri *recInstance) Types() {
	if !ri.Merge.InputType().LTE(ri.inputType.At("out")) {
		ri.Merge.InputType().Meets(ri.inputType.At("out"))
		ri.ToType["merge"] = struct{}{}
	}
	if !ri.Split.OutputType().LTE(ri.outputType.At("out")) {
		ri.Split.OutputType().Meets(ri.outputType.At("out"))
		ri.ToType["split"] = struct{}{}
	}

	for typeNext := ""; len(ri.ToType) > 0; {
		for typeNext = range ri.ToType {
			break
		}
		switch delete(ri.ToType, typeNext); typeNext {
		case "merge":
			recTypeBefore := (*ri.Merge.InputType()).Fields["rec"]
			innerInputTypeBefore := *ri.Merge.OutputType()

			ri.Merge.Types()
			(*ri.Split.OutputType()).Fields["rec"] = (*ri.Merge.InputType()).Fields["rec"]

			if !recTypeBefore.LTE((*ri.Merge.InputType()).Fields["rec"]) {
				ri.Split.OutputType().Meets((*ri.Merge.InputType()).Fields["rec"].At("rec"))
				ri.ToType["split"] = struct{}{}
			}
			if !innerInputTypeBefore.LTE(*ri.Merge.OutputType()) {
				ri.Inner.InputType().Meets(*ri.Merge.OutputType())
				ri.ToType["inner"] = struct{}{}
			}
		case "split":
			recTypeBefore := (*ri.Split.OutputType()).Fields["rec"]
			innerOutputTypeBefore := *ri.Split.InputType()

			ri.Split.Types()
			(*ri.Merge.InputType()).Fields["rec"] = (*ri.Split.OutputType()).Fields["rec"]

			if !recTypeBefore.LTE((*ri.Split.OutputType()).Fields["rec"]) {
				ri.Merge.InputType().Meets((*ri.Split.OutputType()).Fields["rec"].At("rec"))
				ri.ToType["merge"] = struct{}{}
			}
			if !innerOutputTypeBefore.LTE(*ri.Split.InputType()) {
				ri.Inner.OutputType().Meets(*ri.Split.InputType())
				ri.ToType["inner"] = struct{}{}
			}
		case "inner":
			innerInputTypeBefore := *ri.Inner.InputType()
			innerOutputTypeBefore := *ri.Inner.OutputType()

			ri.Inner.Types()

			if !innerInputTypeBefore.LTE(*ri.Inner.InputType()) {
				ri.Merge.OutputType().Meets(*ri.Inner.InputType())
				ri.ToType["merge"] = struct{}{}
			}
			if !innerOutputTypeBefore.LTE(*ri.Inner.OutputType()) {
				ri.Split.InputType().Meets(*ri.Inner.OutputType())
				ri.ToType["split"] = struct{}{}
			}
		}
	}

	ri.inputType = (*ri.Merge.InputType()).Fields["out"]
	ri.outputType = (*ri.Split.OutputType()).Fields["out"]

	if ri.inputType.LTE(NullType) || ri.outputType.LTE(NullType) {
		ri.inputType = NullType
		ri.outputType = NullType
		ri.Merge.InputType().Meets(NullType)
		ri.Split.InputType().Meets(NullType)
		ri.Inner.InputType().Meets(NullType)
		ri.Merge.Types()
		ri.Split.Types()
		ri.Inner.Types()
	}
}

func (ri recInstance) Underdefined() (underdefined Error) {
	underdefined.Ors(
		ri.Merge.Underdefined().Context("in rec merge")).Ors(
		ri.Split.Underdefined().Context("in rec split")).Ors(
		ri.Inner.Underdefined().Context("in rec inner"))
	return
}

func (ri recInstance) Run(input Input, output Output) {
	innerInput, mergeOutput := MakeIO(*ri.Merge.OutputType())
	splitInput, innerOutput := MakeIO(*ri.Split.InputType())
	recInput, recOutput := MakeIO((*ri.Merge.InputType()).Fields["rec"])
	splitOutput := Output{Fields: make(map[string]Output, 2)}
	splitOutput.Fields["rec"] = recOutput
	splitOutput.Fields["out"] = output
	mergeInput := Input{Fields: make(map[string]Input, 2)}
	mergeInput.Fields["out"] = input
	mergeInput.Fields["rec"] = recInput

	go ri.Inner.Run(innerInput, innerOutput)
	go ri.Merge.Run(mergeInput, mergeOutput)
	go ri.Split.Run(splitInput, splitOutput)
}
