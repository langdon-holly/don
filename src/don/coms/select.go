package coms

import . "don/core"

type SelectCom string

func (sc SelectCom) Types(inputType, outputType *DType) (underdefined Error) {
	outputType.Meets(inputType.Get(string(sc)))
	scInputType := MakeNStructType(1)
	scInputType.Fields[string(sc)] = *outputType
	inputType.Meets(scInputType)
	return outputType.Underdefined().Context("in output from select field " + string(sc))
}
func (sc SelectCom) Run(inputType, outputType DType, input Input, output Output) {
	if len(inputType.Fields) > 0 {
		ICom{}.Run(inputType.Fields[string(sc)], outputType, input.Fields[string(sc)], output)
	}
}
