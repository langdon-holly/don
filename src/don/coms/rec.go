package coms

import . "don/core"

type RecCom struct{ Inner Com }

var recComInOutType = MakeNStructType(2)

func init() {
	recComInOutType.Fields["rec"] = UnknownType
}

func (rc RecCom) types(inputType, outputType *DType) (done bool, innerInputType, innerOutputType, mergeInputType, splitOutputType DType) {
	mergeInputType, splitOutputType = recComInOutType, recComInOutType
	mergeInputType.RemakeFields()
	mergeInputType.Fields["out"] = *inputType
	splitOutputType.RemakeFields()
	splitOutputType.Fields["out"] = *outputType

	mergeDone, splitDone, innerDone := false, false, false

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
			recTypeBefore := mergeInputType.Fields["rec"]
			innerInputTypeBefore := innerInputType

			mergeDone = MergeCom{}.Types(&mergeInputType, &innerInputType)
			splitOutputType.Fields["rec"] = mergeInputType.Fields["rec"]

			if !recTypeBefore.LTE(mergeInputType.Fields["rec"]) {
				toType["split"] = struct{}{}
			}
			if !innerInputTypeBefore.LTE(innerInputType) {
				toType["inner"] = struct{}{}
			}
		case "split":
			recTypeBefore := splitOutputType.Fields["rec"]
			innerOutputTypeBefore := innerOutputType

			splitDone = SplitCom{}.Types(&innerOutputType, &splitOutputType)
			mergeInputType.Fields["rec"] = splitOutputType.Fields["rec"]

			if !recTypeBefore.LTE(splitOutputType.Fields["rec"]) {
				toType["merge"] = struct{}{}
			}
			if !innerOutputTypeBefore.LTE(innerOutputType) {
				toType["inner"] = struct{}{}
			}
		case "inner":
			innerInputTypeBefore := innerInputType
			innerOutputTypeBefore := innerOutputType

			innerDone = rc.Inner.Types(&innerInputType, &innerOutputType)

			if !innerInputTypeBefore.LTE(innerInputType) {
				toType["merge"] = struct{}{}
			}
			if !innerOutputTypeBefore.LTE(innerOutputType) {
				toType["split"] = struct{}{}
			}
		}
	}

	*inputType = mergeInputType.Fields["out"]
	*outputType = splitOutputType.Fields["out"]
	done = mergeDone && splitDone && innerDone
	return
}

// Violates multiplicative annihilation!!
func (rc RecCom) Types(inputType, outputType *DType) (done bool) {
	done, _, _, _, _ = rc.types(inputType, outputType)
	return
}

func (rc RecCom) Run(inputType, outputType DType, input Input, output Output) {
	_, innerInputType, innerOutputType, mergeInputType, splitOutputType := rc.types(&inputType, &outputType)

	innerInput, mergeOutput := MakeIO(innerInputType)
	splitInput, innerOutput := MakeIO(innerOutputType)
	recInput, recOutput := MakeIO(mergeInputType.Fields["rec"])
	splitOutput := Output{Fields: make(map[string]Output, 2)}
	splitOutput.Fields["rec"] = recOutput
	splitOutput.Fields["out"] = output
	mergeInput := Input{Fields: make(map[string]Input, 2)}
	mergeInput.Fields["out"] = input
	mergeInput.Fields["rec"] = recInput

	go rc.Inner.Run(innerInputType, innerOutputType, innerInput, innerOutput)
	go MergeCom{}.Run(mergeInputType, innerInputType, mergeInput, mergeOutput)
	go SplitCom{}.Run(innerOutputType, splitOutputType, splitInput, splitOutput)
}
