package syntax

type SyntaxTag int

const (
	ListSyntaxTag      = SyntaxTag(iota)
	EmptyLineSyntaxTag /* child only of list */
	SpacedSyntaxTag
	MCallSyntaxTag
	SandwichSyntaxTag
	NameSyntaxTag
)

type Syntax struct {
	Tag SyntaxTag

	// for Tag == ListSyntaxTag ||
	//  Tag == SpacedSyntaxTag ||
	//  Tag == MCallSyntaxTag ||
	//  Tag == SandwichSyntaxTag
	// Nonempty for SpacedSyntaxTag
	// 2 elements for MCallSyntaxTag or SandwichSyntaxTag
	Children []Syntax

	// for Tag == NameSyntaxTag
	LeftMarker, RightMarker bool
	Name                    string
}
