package coms

import . "don/core"

type DeselectCom string

func (dc DeselectCom) Instantiate() ComInstance {
	return &deselectInstance{FieldName: string(dc)}
}

func (dc DeselectCom) Inverse() Com { return SelectCom(string(dc)) }

type deselectInstance struct {
	FieldName             string
	inputType, outputType DType
}

func (di *deselectInstance) InputType() *DType  { return &di.inputType }
func (di *deselectInstance) OutputType() *DType { return &di.outputType }
func (di *deselectInstance) Types() {
	di.inputType.Meets(di.outputType.Get(di.FieldName))
	diOutputType := MakeNFieldsType(1)
	diOutputType.Fields[di.FieldName] = di.inputType
	di.outputType.Meets(diOutputType)
}
func (di deselectInstance) Underdefined() Error {
	return di.inputType.Underdefined().Context(
		"in input to deselect field " + di.FieldName)
}
func (di deselectInstance) Run(input Input, output Output) {
	if len(di.outputType.Fields) > 0 {
		RunI(di.inputType, input, output.Fields[di.FieldName])
	}
}
