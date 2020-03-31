package core

type Com interface {
	OutputType(inputType DType) (outputType DType, impossible bool)
	Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{})
}
