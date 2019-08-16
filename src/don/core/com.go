package core

type Com interface {
	Run(input interface{}, output interface{}, quit <-chan struct{})
	InputType() DType
	OutputType() DType
}
