package coms

import . "don/core"

type SelectCom string

func (sc SelectCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	done = true
	*outputType, bad = MergeTypes(*outputType, inputType.Fields[string(sc)])
	if bad != nil {
		bad = append(bad, "in bad output type for select :"+string(sc))
		return
	}
	scInputType := MakeNStructType(1)
	scInputType.Fields[string(sc)] = *outputType
	*inputType, bad = MergeTypes(*inputType, scInputType)
	if bad != nil {
		bad = append(bad, "in bad input type for select :"+string(sc))
	}
	return
}

func (sc SelectCom) Run(inputType, outputType DType, input Input, output Output) {
	ICom{}.Run(inputType.Fields[string(sc)], outputType, input.Fields[string(sc)], output)
}
