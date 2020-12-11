package syntax

const (
	tab        byte = 9
	lf         byte = 10
	space      byte = 32
	bang       byte = 33
	hash       byte = 35
	dollar     byte = 36
	leftParen  byte = 40
	rightParen byte = 41
	hyphen     byte = 45
	colon      byte = 58
	backslash  byte = 92
	underscore byte = 95
)

func byteIsSpecial(b byte) bool {
	return b == tab ||
		b == lf ||
		b == space ||
		b == bang ||
		b == hash ||
		b == dollar ||
		b == leftParen ||
		b == rightParen ||
		b == hyphen ||
		b == colon ||
		b == backslash ||
		b == underscore
}
