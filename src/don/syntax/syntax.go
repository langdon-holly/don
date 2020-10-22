package syntax

type SyntaxTag int

const (
	ListSyntaxTag = SyntaxTag(iota)
	SpacedSyntaxTag
	MCallSyntaxTag
	NameSyntaxTag
)

type Syntax struct {
	Tag SyntaxTag

	Children                []Syntax /* for Tag == ListSyntaxTag || Tag == SpacedSyntaxTag */
	LeftMarker, RightMarker bool     /* for Tag == ListSyntaxTag || Tag == NameSyntaxTag || Tag == MCallSyntaxTag */
	Name                    string   /* for Tag == NameSyntaxTag || Tag == MCallSyntaxTag */
	Child                   *Syntax  /* for Tag == MCallSyntaxTag */
}
