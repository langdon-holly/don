package core

type Com interface {
	Types(inputType, outputType *DType) (bad []string, done bool)
	Run(inputType, outputType DType, input Input, output Output)
}
