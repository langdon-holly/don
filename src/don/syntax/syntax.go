package syntax

type SyntaxTag int

const (
	ListSyntaxTag      = SyntaxTag(iota)
	EmptyLineSyntaxTag /* child only of list */
	ApplicationSyntaxTag
	CompositionSyntaxTag
	NameSyntaxTag
)

type Syntax struct {
	Tag SyntaxTag

	// for Tag == ListSyntaxTag ||
	//  Tag == ApplicationSyntaxTag ||
	//  Tag == CompositionSyntaxTag ||
	// 2 elements for ApplicationSyntaxTag
	// Nonempty for CompositionSyntaxTag
	Children []Syntax

	// for Tag == NameSyntaxTag
	LeftMarker, RightMarker bool
	Name                    string
}
