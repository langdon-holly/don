package syntax

const ( /* Order matters */
	DisjunctionPrecedence = iota
	ConjunctionPrecedence
	EmptyLinePrecedence
	LeftAssociativePrecedence
	CompositionPrecedence
	NamedPrecedence
	ISyntaxPrecedence
	QuotePrecedence
)

type Syntax interface {
	precedence() int
	layout() (l layoutInfo, ws writeString)
	String() string
}

type Disjunction struct{ Disjuncts []Syntax }
type Conjunction struct{ Conjuncts []Syntax }
type EmptyLine struct{} /* Only in disjunction or conjunction; neither first nor last */
type Application struct{ Com, Arg Syntax }
type Bind struct{ Body, Var Syntax }
type Composition struct {
	Factors []Syntax /* Nonempty */
}
type Named struct {
	LeftMarker, RightMarker bool
	Name                    string
}
type ISyntax struct{}
type Quote struct{ Syntax Syntax }
