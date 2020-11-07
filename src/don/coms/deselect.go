package coms

import . "don/core"

type DeselectCom string

func (dc DeselectCom) Types(inputType, outputType *DType) (underdefined Error) {
	inputType.Meets(outputType.Get(string(dc)))
	dcOutputType := MakeNStructType(1)
	dcOutputType.Fields[string(dc)] = *inputType
	outputType.Meets(dcOutputType)
	return inputType.Underdefined().Context("in input to deselect field " + string(dc))
}
func (dc DeselectCom) Run(inputType, outputType DType, input Input, output Output) {
	if len(outputType.Fields) > 0 {
		ICom{}.Run(inputType, outputType.Fields[string(dc)], input, output.Fields[string(dc)])
	}
}
