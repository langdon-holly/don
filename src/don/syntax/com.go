package syntax

import (
	"don/coms"
	. "don/core"
)

func (s Syntax) ToCom() Com {
	switch s.Tag {
	case BlockSyntaxTag:
		pipes := make([]Com, len(s.Children))
		for i, line := range s.Children {
			pipeComs := make([]Com, len(line))
			for j, subS := range line {
				pipeComs[len(pipeComs)-1-j] = subS.ToCom()
			}
			pipes[i] = coms.Pipe(pipeComs)
		}
		return coms.SplitMerge(pipes)
	case MCallSyntaxTag:
		switch s.Name {
		case "com":
		case "prod":
		}
		panic("Unknown macro")
	case MacroSyntaxTag:
		switch s.Name {
		case "init":
			return coms.InitCom{}
		}
		panic("Unknown macro")
	case SelectSyntaxTag:
		return coms.SelectCom(s.Name)
	case DeselectSyntaxTag:
		return coms.Deselect(s.Name)
	case IsolateSyntaxTag:
		return coms.Pipe([]Com{coms.SelectCom(s.Name), coms.Deselect(s.Name)})
	}
	panic("Unreachable")
}
