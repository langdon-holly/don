package core

type Com interface {
	InputType() DType
	OutputType() DType
	MeetTypes(inputType, outputType DType) Com /* Invalidates */
	Underdefined() Error
	Copy() Com
	Invert() Com /* Invalidates */
	TypedCom(tcb TypedComBuilder /* mutated */, inputMap, outputMap TypeMap)
}
