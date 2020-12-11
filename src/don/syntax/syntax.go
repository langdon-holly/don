package syntax

type SyntaxTag int

const (
	ListSyntaxTag = SyntaxTag(iota)
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
	// 1 element for MCallSyntaxTag; 2 elements for SandwichSyntaxTag
	Children []Syntax

	LeftMarker, RightMarker bool   /* for Tag == NameSyntaxTag || Tag == MCallSyntaxTag */
	Name                    string /* for Tag == NameSyntaxTag || Tag == MCallSyntaxTag */
}
