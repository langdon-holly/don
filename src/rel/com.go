package rel

type Rel interface {
	Type() *TypePtr
	Copy(map[*TypePtr]*TypePtr /* mutated */) Rel
	Convert() Rel /* Invalidates */
	Syntax() Syntax
	String() string
}
