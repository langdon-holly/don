package syntax

import "strings"

const ( /* Order matters */
	ListPrecedence = iota
	EmptyLinePrecedence
	ApplicationPrecedence
	CompositionPrecedence
	NamedPrecedence
	ISyntaxPrecedence
	QuotePrecedence
)

type Syntax interface {
	precedence() int
	layout() (l layoutInfo, writeString func(out *strings.Builder, indent []byte))
	String() string
}

type List struct{ Factors []Syntax }
type EmptyLine struct{} /* Only in list; neither first nor last factor */
type Application struct{ Com, Arg Syntax }
type Composition struct {
	Factors []Syntax /* Nonempty */
}
type Named struct {
	LeftMarker, RightMarker bool
	Name                    string
}
type ISyntax struct{}
type Quote struct{ Syntax Syntax }
