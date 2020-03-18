package syntax

import "strconv"

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
				pipeComs[len(line)-1-j] = subS.ToCom()
			}
			pipes[i] = coms.Pipe(pipeComs)
		}
		return coms.SplitMerge(pipes)
	case MCallSyntaxTag:
		switch s.Name {
		case "com":
			if len(s.Children) < 1 {
				panic("Empty [com] body")
			}

			pipes := make([]Com, len(s.Children))
			for i, line := range s.Children {
				pipeComs := make([]Com, len(line))
				for j, subS := range line {
					pipeComs[len(line)-1-j] = subS.ToCom()
				}
				pipes[i] = coms.Pipe(pipeComs)
			}
			return coms.ComCom(pipes)
		case "prod":
			pipes := make([]Com, len(s.Children))
			for i, line := range s.Children {
				pipeComs := make([]Com, len(line)+1)
				for j, subS := range line {
					pipeComs[len(line)-1-j] = subS.ToCom()
				}
				pipeComs[len(line)] = coms.Deselect(strconv.FormatInt(int64(i), 10))
				pipes[i] = coms.Pipe(pipeComs)
			}
			return coms.Pipe([]Com{coms.SplitMerge(pipes), coms.ProdCom{}})
		}
		panic("Unknown macro")
	case MacroSyntaxTag:
		switch s.Name {
		case "I":
			return coms.ICom{}
		case "init":
			return coms.InitCom{}
		}
		panic("Unknown macro")
	case SelectSyntaxTag:
		return coms.SelectCom(s.Name)
	case DeselectSyntaxTag:
		return coms.Deselect(s.Name)
	}
	panic("Unreachable")
}
