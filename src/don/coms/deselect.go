package coms

import . "don/core"

type DeselectCom string

func (dc DeselectCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	done = true

	if bad = inputType.Meets(outputType.Fields[string(dc)]); bad != nil {
		bad = append(bad, "in bad input type for deselect "+string(dc)+":")
		return
	}
	dcOutputType := MakeNStructType(1)
	dcOutputType.Fields[string(dc)] = *inputType
	if bad = outputType.Meets(dcOutputType); bad != nil {
		bad = append(bad, "in bad output type for deselect "+string(dc)+":")
	}
	return
}
func (dc DeselectCom) Run(inputType, outputType DType, input Input, output Output) {
	ICom{}.Run(inputType, outputType.Fields[string(dc)], input, output.Fields[string(dc)])
}
