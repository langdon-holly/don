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
		Merge: MergeCom{}.Instantiate(),
		Split: SplitCom{}.Instantiate(),
		Inner: rc.Inner.Instantiate()}
	ri.Merge.InputType().Meets(recComInOutType)
	ri.Split.OutputType().Meets(recComInOutType)
	return &ri
}

type recInstance struct {
	Merge, Split, Inner   ComInstance
	inputType, outputType DType
}

func (ri *recInstance) InputType() *DType  { return &ri.inputType }
func (ri *recInstance) OutputType() *DType { return &ri.outputType }

// Violates multiplicative annihilation!!
func (ri *recInstance) Types() {
	ri.Merge.InputType().MeetsAtPath(ri.inputType, []string{"out"})
	ri.Split.OutputType().MeetsAtPath(ri.outputType, []string{"out"})

	toType := make(map[string]struct{}, 3)
	toType["merge"] = struct{}{}
	toType["split"] = struct{}{}
	toType["inner"] = struct{}{}
	for len(toType) > 0 {
		var typeNext string
		for typeNext = range toType {
			delete(toType, typeNext)
			break
		}
		switch typeNext {
		case "merge":
			recTypeBefore := (*ri.Merge.InputType()).Fields["rec"]
			innerInputTypeBefore := *ri.Merge.OutputType()

			ri.Merge.Types()
			(*ri.Split.OutputType()).Fields["rec"] = (*ri.Merge.InputType()).Fields["rec"]

			if !recTypeBefore.LTE((*ri.Merge.InputType()).Fields["rec"]) {
				ri.Split.OutputType().MeetsAtPath((*ri.Merge.InputType()).Fields["rec"], []string{"rec"})
				toType["split"] = struct{}{}
			}
			if !innerInputTypeBefore.LTE(*ri.Merge.OutputType()) {
				ri.Inner.InputType().Meets(*ri.Merge.OutputType())
				toType["inner"] = struct{}{}
			}
		case "split":
			recTypeBefore := (*ri.Split.OutputType()).Fields["rec"]
			innerOutputTypeBefore := *ri.Split.InputType()

			ri.Split.Types()
			(*ri.Merge.InputType()).Fields["rec"] = (*ri.Split.OutputType()).Fields["rec"]

			if !recTypeBefore.LTE((*ri.Split.OutputType()).Fields["rec"]) {
				ri.Merge.InputType().MeetsAtPath((*ri.Split.OutputType()).Fields["rec"], []string{"rec"})
				toType["merge"] = struct{}{}
			}
			if !innerOutputTypeBefore.LTE(*ri.Split.InputType()) {
				ri.Inner.OutputType().Meets(*ri.Split.InputType())
				toType["inner"] = struct{}{}
			}
		case "inner":
			innerInputTypeBefore := *ri.Inner.InputType()
			innerOutputTypeBefore := *ri.Inner.OutputType()

			ri.Inner.Types()

			if !innerInputTypeBefore.LTE(*ri.Inner.InputType()) {
				ri.Merge.OutputType().Meets(*ri.Inner.InputType())
				toType["merge"] = struct{}{}
			}
			if !innerOutputTypeBefore.LTE(*ri.Inner.OutputType()) {
				ri.Split.InputType().Meets(*ri.Inner.OutputType())
				toType["split"] = struct{}{}
			}
		}
	}

	ri.inputType = (*ri.Merge.InputType()).Fields["out"]
	ri.outputType = (*ri.Split.OutputType()).Fields["out"]
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
