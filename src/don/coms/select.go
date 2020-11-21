package coms

import . "don/core"

type SelectCom string

func (sc SelectCom) Instantiate() ComInstance {
	return &selectInstance{FieldName: string(sc)}
}

type selectInstance struct {
	FieldName             string
	inputType, outputType DType
}

func (si *selectInstance) InputType() *DType  { return &si.inputType }
func (si *selectInstance) OutputType() *DType { return &si.outputType }
func (si *selectInstance) Types() (underdefined Error) {
	si.outputType.Meets(si.inputType.Get(si.FieldName))
	siInputType := MakeNStructType(1)
	siInputType.Fields[si.FieldName] = si.outputType
	si.inputType.Meets(siInputType)
	return si.outputType.Underdefined().Context("in output from select field " + si.FieldName)
}
func (si selectInstance) Run(input Input, output Output) {
	if len(si.inputType.Fields) > 0 {
		RunI(si.outputType, input.Fields[si.FieldName], output)
	}
}
