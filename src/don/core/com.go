package core

type Com interface {
	OutputType(inputType PartialType) PartialType
	Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{})
}
