package core

type Com interface {
	OutputType(inputType DType) DType
	Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{})
}
