package com

import (
	"os"
	"path"
)

import (
	. "don/junctive"
	"don/syntax"
)

type Comment struct{}
type EvalResult struct{ It interface{} }

func (r EvalResult) Com() Com             { return r.It.(Com) }
func (r EvalResult) Syntax() syntax.Words { return r.It.(syntax.Words) }
func (r EvalResult) Apply(param EvalResult) EvalResult {
	return EvalResult{r.It.(func(EvalResult) interface{})(param)}
}
func (r EvalResult) IsComment() bool { _, isComment := r.It.(Comment); return isComment }

type context *struct {
	Entries map[string]interface{}
	Parent  context
	Dir     string
}

var defContext = new(struct {
	Entries map[string]interface{}
	Parent  context
	Dir     string
})

func init() {
	defContext.Entries = make(map[string]interface{})

	defContext.Entries["I"] = Composition(nil)
	defContext.Entries["false"] = Junction(ConJunctive, nil)
	defContext.Entries["true"] = Junction(DisJunctive, nil)
	//defContext.Entries["swap"] = Swap()
	defContext.Entries["~"] = func(param EvalResult) interface{} {
		return param.Com().Convert()
	}
	//defContext.Entries["debug-type"] = DebugType
	//defContext.Entries["sandwich"] = func(param0 EvalResult) interface{} {
	//	return func(param1 EvalResult) interface{} {
	//		return Composition([]Com{
	//			param0.Com(),
	//			param1.Com(),
	//			param0.Com().Copy(make(map[*TypePtr]*TypePtr)).Convert(),
	//		})
	//	}
	//}
}

func pathJoin(dir, file string) string { return dir + "/" + file }

// c may be shared
func evalWord(w syntax.Word, c context) interface{} {
	if len(w.Specials) == 0 {
		name := w.Strings[0]
		for cNow := c; cNow != nil; cNow = cNow.Parent {
			if entry, ok := cNow.Entries[name]; !ok {
			} else if com, isCom := entry.(Com); isCom {
				return com.Copy(make(map[*TypePtr]*TypePtr))
			} else {
				return entry
			}
		}

		dir := ""
		for cNow := c; dir == ""; cNow = cNow.Parent {
			if cNow == nil {
				panic("No filesystem")
			}
			dir = cNow.Dir
		}
		val := EvalFile(pathJoin(dir, name)).It
		if c.Entries == nil {
			c.Entries = make(map[string]interface{}, 1)
		}
		c.Entries[name] = val
		if com, isCom := val.(Com); isCom {
			return com.Copy(make(map[*TypePtr]*TypePtr))
		} else {
			return val
		}
	} else {
		switch specialPayload := w.Specials[0].(type) {
		case syntax.WordSpecialDelimited:
			if len(w.Specials) > 1 {
				panic(
					"Overly special word: " + w.String(),
				)
			}
			if w.Strings[0] != "" || w.Strings[1] != "" {
				panic("Delimitation embedded in name: " + w.String())
			}
			if specialPayload.LeftDelim != specialPayload.RightDelim {
				panic(
					"Unmatched delimiters: " +
						specialPayload.LeftDelim.String() +
						" and " +
						specialPayload.RightDelim.String(),
				)
			}
			switch specialPayload.LeftDelim {
			case syntax.MaybeDelimNone:
				panic("Missing delimiters")
			case syntax.MaybeDelimParen:
				return evalWords(specialPayload.Words, c)
			case syntax.MaybeDelimBrace:
				return specialPayload.Words
			}
			panic("Unreachable")
		case syntax.WordSpecialJunct:
			if 1 < len(w.Specials) {
				panic(
					"Overly special word: " + w.String(),
				)
			}
			if w.Strings[0] == "" {
				if w.Strings[1] == "" {
					panic("Junct with no name: " + w.String())
				} else {
					return Collect(Junctive(specialPayload), w.Strings[1])
				}
			} else {
				if w.Strings[1] == "" {
					return Select(Junctive(specialPayload), w.Strings[0])
				} else {
					panic("Junct with two names: " + w.String())
				}
			}
		case syntax.WordSpecialCommentMarker:
			if w.Strings[0] != "" {
				panic("Named comment: " + w.String())
			}
			return Comment{}
		}
		panic("Unreachable")
	}
}

// c may be shared
func evalComposition(composition []syntax.Word, c context) interface{} {
	var factorResults []EvalResult // No Comments
	for _, factor := range composition {
		if er := (EvalResult{evalWord(factor, c)}); !er.IsComment() {
			factorResults = append(factorResults, er)
		}
	}
	if len(factorResults) == 0 {
		panic("Empty composition")
	}
	if _, macrosp := factorResults[0].It.(func(EvalResult) interface{}); macrosp {
		return func(param EvalResult) interface{} {
			for _, factorResult := range factorResults {
				param = factorResult.Apply(param)
			}
			return param.It
		}
	} else {
		factorComs := make([]Com, len(factorResults))
		for i, factorResult := range factorResults {
			factorComs[i] = factorResult.Com()
		}
		return Composition(factorComs)
	}
}

// c may be shared
func EvalComposition(composition []syntax.Word, c context) EvalResult {
	return EvalResult{evalComposition(composition, c)}
}

// c may be shared
func evalWords(ws syntax.Words, c context) interface{} {
	if 0 < len(ws.Operators) {
		for _, firstSpecial := range ws.Operators[0].Specials {
			switch firstSpecialPayload := firstSpecial.(type) {
			case syntax.WordSpecialJunction:
				if len(ws.Compositions[0]) != 0 {
					panic("Junction doesn't start with operator word")
				}
				var junctComs []Com
				for i, operator := range ws.Operators {
					// 0 < len(operator.Specials)
					_, commented := operator.Specials[0].(syntax.WordSpecialCommentMarker)
					if commented {
						if operator.Strings[0] != "" {
							panic("Bad junction operator word: " + operator.String())
						}
						// There is at least one operator special in `operator`, but a comment marker isn't one;
						// therefore, 1 < len(operator.Specials)
						operator = syntax.Word{Strings: operator.Strings[1:], Specials: operator.Specials[1:]}
					}
					// 0 < len(operator.Specials)
					if operator.Strings[0] != "" || operator.Strings[1] != "" || 1 < len(operator.Specials) {
						panic("Bad junction operator word: " + operator.String())
					}
					if specialPayload, isWordSpecialJunction :=
						operator.Specials[0].(syntax.WordSpecialJunction); !isWordSpecialJunction ||
						specialPayload != firstSpecialPayload {
						panic("Bad junction operator word: " + operator.String())
					}

					if !commented {
						junctComs = append(junctComs, EvalComposition(ws.Compositions[1:][i], c).Com())
					}
				}
				if len(junctComs) == 0 {
					panic("Empty junction: " + ws.String())
				}
				return Junction(Junctive(firstSpecialPayload), junctComs)
			case syntax.WordSpecialApplication:
				val := EvalComposition(ws.Compositions[0], c)
				for i, operator := range ws.Operators {
					if len(operator.Specials) != 1 ||
						operator.Strings[0] != "" ||
						operator.Strings[1] != "" {
						panic("Bad application operator word: " + operator.String())
					}
					if _, isWordSpecialApplication :=
						operator.Specials[0].(syntax.WordSpecialApplication); !isWordSpecialApplication {
						panic("Bad application operator word: " + operator.String())
					}

					val = val.Apply(EvalComposition(ws.Compositions[1:][i], c))
				}
				return val.It
			}
		}
		panic("Unreachable")
	} else {
		return evalComposition(ws.Compositions[0], c)
	}
}

func EvalFile(srcPath string) EvalResult {
	if file, err := os.Open(srcPath); err == nil {
		//for _, p := range []string{"/home/lzh", "/home/lzh/", "/", ".", "..", "don"} {
		//	dir, file := path.Split(p)
		//	fmt.Printf("%s<\t>%s\n", dir, file)
		//}

		for {
			if dest, err := os.Readlink(srcPath); err != nil {
				break
			} else if dir, _ := path.Split(srcPath); true {
				srcPath = pathJoin(dir, dest)
			}
		}
		dir, _ := path.Split(srcPath)
		if dir == "" {
			dir = "."
		}

		return EvalResult{evalWords(syntax.Parse(file), &struct {
			Entries map[string]interface{}
			Parent  context
			Dir     string
		}{Entries: nil, Parent: defContext, Dir: dir})}
	} else {
		panic(err)
	}
}
