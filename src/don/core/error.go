package core

import "strings"

import "don/syntax"

// nil or nonempty
type Error []string

func NewError(msg string) Error { return Error([]string{msg}) }

// Sets e
func (e *Error) Remake() { *e = append(Error(nil), *e...) }

// Mutates e
func (e Error) Context(msg string) Error {
	if e == nil {
		return e
	} else {
		return append(e, msg)
	}
}
func (e Error) InField(fieldName string) Error {
	return e.Context("in field " + syntax.EscapeFieldName(fieldName))
}

func (e0 *Error) Ors(e1 Error) *Error {
	if *e0 == nil {
		*e0 = e1
	}
	return e0
}

func (e Error) String() string {
	if e == nil {
		return ""
	} else {
		var b strings.Builder
		b.WriteString(e[0])
		for _, msg := range e[1:] {
			b.WriteString("\n")
			b.WriteString(msg)
		}
		return b.String()
	}
}
