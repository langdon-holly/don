package core

type Com interface {
	InputType() *DType  /* Invalidated by Types or Invert; may alias OutputType(); after mutating, call Types */
	OutputType() *DType /* Invalidated by Types or Invert; may alias InputType(); after mutating, call Types */
	Types() Com         /* Invalidates */
	Underdefined() Error
	Copy() Com
	Invert() Com /* Invalidates */
	Run(input Input, output Output)
}
