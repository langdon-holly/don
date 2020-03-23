package syntax

type SyntaxTag int

const (
	BindSyntaxTag = SyntaxTag(iota)
	BlockSyntaxTag
	MCallSyntaxTag
	MacroSyntaxTag
	SelectSyntaxTag
	DeselectSyntaxTag
)

type Syntax struct {
	Tag SyntaxTag

	Name            string     /* for Tag != BindSyntaxTag && Tag != BlockSyntaxTag */
	LeftAt, RightAt bool       /* for Tag == BlockSyntaxTag */
	Children        [][]Syntax /* for Tag == BindSyntaxTag || Tag == BlockSyntaxTag || Tag == MCallSyntaxTag*/
}
