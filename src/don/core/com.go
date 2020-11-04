package core

type Com interface {
	Types(inputType, outputType *DType) (done bool)
	Run(inputType, outputType DType, input Input, output Output)
}
