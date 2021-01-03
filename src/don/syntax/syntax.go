package syntax

type SyntaxTag int

const (
	ListSyntaxTag      = SyntaxTag(iota)
	EmptyLineSyntaxTag /* child only of list */
	CompositionSyntaxTag
	ApplicationSyntaxTag
	SandwichSyntaxTag
	NameSyntaxTag
)

type Syntax struct {
	Tag SyntaxTag

	// for Tag == ListSyntaxTag ||
	//  Tag == CompositionSyntaxTag ||
	//  Tag == ApplicationSyntaxTag ||
	//  Tag == SandwichSyntaxTag
	// Nonempty for CompositionSyntaxTag
	// 2 elements for ApplicationSyntaxTag or SandwichSyntaxTag
	Children []Syntax

	// for Tag == NameSyntaxTag
	LeftMarker, RightMarker bool
	Name                    string
}
