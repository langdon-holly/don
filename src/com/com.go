package com

type Com interface {
	Type() *TypePtr
	Copy(map[*TypePtr]*TypePtr /* mutated */) Com
	Convert() Com /* Invalidates */
	Syntax() Syntax
	String() string
}
