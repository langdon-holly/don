package syntax

import (
	"io"
	"strings"
)

func nextByte(input io.Reader) (byte, bool) {
	var bs [1]byte
	for {
		n, err := input.Read(bs[:])
		if n == 1 {
			return bs[0], false
		}
		if err != nil {
			return 0, true
		}
	}
}

type input struct {
	Reader   io.Reader
	Buffered bool
	NextByte byte
	EOF      bool
}

func (in *input) Next() (byte, bool) {
	if in.Buffered {
		in.Buffered = false
		return in.NextByte, in.EOF
	} else {
		return nextByte(in.Reader)
	}
}

func (in *input) Peek() (byte, bool) {
	if !in.Buffered {
		in.Buffered = true
		in.NextByte, in.EOF = nextByte(in.Reader)
	}
	return in.NextByte, in.EOF
}

func parseTop(in *input) [][]Syntax {
	var lines [][]Syntax
	var currentLine []Syntax
	ready := true

	for {
		b, eof := in.Peek()
		if eof {
			in.Next()
			if len(currentLine) > 0 {
				lines = append(lines, currentLine)
			}
			return lines
		}
		switch b {
		case ' ':
			fallthrough
		case '\t':
			in.Next()
			ready = true
		case '\n':
			in.Next()
			if len(currentLine) > 0 {
				lines = append(lines, currentLine)
				currentLine = nil
			}
			ready = true
		default:
			if !ready {
				panic("Syntax error")
			}
			currentLine = append(currentLine, parse(in))
			ready = false
		}
	}
}

func parseChildren(in *input) [][]Syntax {
	var lines [][]Syntax
	var currentLine []Syntax
	ready := true

	for {
		b, eof := in.Peek()
		if eof {
			panic("Syntax error")
		}
		switch b {
		case ' ':
			fallthrough
		case '\t':
			in.Next()
			ready = true
		case '\n':
			in.Next()
			if len(currentLine) > 0 {
				lines = append(lines, currentLine)
				currentLine = nil
			}
			ready = true
		case ')':
			in.Next()
			if len(currentLine) > 0 {
				lines = append(lines, currentLine)
			}
			return lines
		default:
			if !ready {
				panic("Syntax error")
			}
			currentLine = append(currentLine, parse(in))
			ready = false
		}
	}
}

func parseName(in *input) string {
	var builder strings.Builder
	depth := 1
	for {
		b, eof := in.Next()
		if eof {
			panic("Syntax error")
		}
		switch b {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return builder.String()
			}
		}
		builder.WriteByte(b)
	}
}

func parse(in *input) Syntax {
	b, _ := in.Next()
	switch b {
	case '[':
		name := parseName(in)
		switch b, _ := in.Peek(); b {
		case ':':
			in.Next()
			return Syntax{Tag: DeselectSyntaxTag, Name: name}
		case '(':
			in.Next()
			return Syntax{Tag: MCallSyntaxTag, Name: name, Children: parseChildren(in)}
		default:
			return Syntax{Tag: MacroSyntaxTag, Name: name}
		}
	case ':':
		if b, _ := in.Next(); b != '[' {
			panic("Syntax error")
		}
		name := parseName(in)
		switch b, _ := in.Peek(); b {
		case ':':
			in.Next()
			return Syntax{Tag: IsolateSyntaxTag, Name: name}
		default:
			return Syntax{Tag: SelectSyntaxTag, Name: name}
		}
	case '(':
		return Syntax{Tag: BlockSyntaxTag, Children: parseChildren(in)}
	}
	panic("Syntax error")
}

func ParseTop(inReader io.Reader) [][]Syntax {
	in := input{Reader: inReader}
	return parseTop(&in)
}
