package coms

import . "don/core"

type SelectCom string

func (sc SelectCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	done = true
	if bad = outputType.Meets(inputType.Fields[string(sc)]); bad != nil {
		bad = append(bad, "in bad output type for select :"+string(sc))
		return
	}
	scInputType := MakeNStructType(1)
	scInputType.Fields[string(sc)] = *outputType
	if bad = inputType.Meets(scInputType); bad != nil {
		bad = append(bad, "in bad input type for select :"+string(sc))
	}
	return
}

func (sc SelectCom) Run(inputType, outputType DType, input Input, output Output) {
	ICom{}.Run(inputType.Fields[string(sc)], outputType, input.Fields[string(sc)], output)
}
