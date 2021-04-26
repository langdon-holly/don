package syntax

const (
	tab        byte = 9
	lf         byte = 10
	space      byte = 32
	bang       byte = 33
	leftParen  byte = 40
	rightParen byte = 41
	asterisk   byte = 42
	colon      byte = 58
	question   byte = 63
	backslash  byte = 92
	underscore byte = 95
	leftBrace  byte = 123
	rightBrace byte = 125
)

func byteIsSpecial(b byte) bool {
	return b == tab ||
		b == lf ||
		b == space ||
		b == bang ||
		b == leftParen ||
		b == rightParen ||
		b == asterisk ||
		b == colon ||
		b == question ||
		b == backslash ||
		b == underscore ||
		b == leftBrace ||
		b == rightBrace
}
