package syntax

const (
	tab        byte = 9
	lf         byte = 10
	space      byte = 32
	bang       byte = 33
	hash       byte = 35
	leftParen  byte = 40
	rightParen byte = 41
	comma      byte = 44
	period     byte = 46
	colon      byte = 58
	semicolon  byte = 59
	leftBrace  byte = 123
	rightBrace byte = 125
)

func byteIsSpecial(b byte) bool {
	return b == tab ||
		b == lf ||
		b == space ||
		b == bang ||
		b == hash ||
		b == leftParen ||
		b == rightParen ||
		b == comma ||
		b == period ||
		b == colon ||
		b == semicolon ||
		b == leftBrace ||
		b == rightBrace
}
