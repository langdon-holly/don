package syntax

type SyntaxTag int

const ( /* Order matters, for printing */
	ListSyntaxTag      = SyntaxTag(iota)
	EmptyLineSyntaxTag /* child only of list */
	ApplicationSyntaxTag
	CompositionSyntaxTag
	NameSyntaxTag
	ISyntaxTag
	QuotationSyntaxTag
)

type Syntax struct {
	Tag SyntaxTag

	// for Tag == ListSyntaxTag ||
	//  Tag == ApplicationSyntaxTag ||
	//  Tag == QuotationSyntaxTag ||
	//  Tag == CompositionSyntaxTag ||
	// 2 elements for ApplicationSyntaxTag
	// 1 element for QuotationSyntaxTag
	// Nonempty for CompositionSyntaxTag
	Children []Syntax

	// for Tag == NameSyntaxTag
	LeftMarker, RightMarker bool
	Name                    string
}
