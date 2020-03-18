package syntax

type SyntaxTag int

const (
	BlockSyntaxTag = SyntaxTag(iota)
	MCallSyntaxTag
	MacroSyntaxTag
	SelectSyntaxTag
	DeselectSyntaxTag
)

type Syntax struct {
	Tag SyntaxTag

	Name     string     /* for Tag != BlockSyntaxTag */
	Children [][]Syntax /* for Tag == BlockSyntaxTag || Tag == MCallSyntaxTag*/
}
