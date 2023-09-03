package syntax

const (
	tab        byte = 9
	lf         byte = 10
	space      byte = 32
	hash       byte = 35
	leftParen  byte = 40
	rightParen byte = 41
	period     byte = 46
	colon      byte = 58
	leftBrace  byte = 123
	pipe       byte = 124
	rightBrace byte = 125
)

func byteIsSpecial(b byte) bool {
	return b == tab ||
		b == lf ||
		b == space ||
		b == hash ||
		b == leftParen ||
		b == rightParen ||
		b == period ||
		b == colon ||
		b == leftBrace ||
		b == pipe ||
		b == rightBrace
}
