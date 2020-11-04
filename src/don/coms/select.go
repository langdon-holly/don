package coms

import . "don/core"

type SelectCom string

func (sc SelectCom) Types(inputType, outputType *DType) (done bool) {
	outputType.Meets(inputType.Get(string(sc)))
	scInputType := MakeNStructType(1)
	scInputType.Fields[string(sc)] = *outputType
	inputType.Meets(scInputType)
	return outputType.Done()
}
func (sc SelectCom) Run(inputType, outputType DType, input Input, output Output) {
	if len(inputType.Fields) > 0 {
		ICom{}.Run(inputType.Fields[string(sc)], outputType, input.Fields[string(sc)], output)
	}
}
