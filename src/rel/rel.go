package rel

type Rel interface {
	Var() *VarPtr
	Copy(map[*VarPtr]*VarPtr /* mutated */, map[*TypePtr]*TypePtr /* mutated */) Rel
	Convert() Rel /* Invalidates */
	Syntax() Syntax
	String() string
}
