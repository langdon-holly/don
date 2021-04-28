package coms

import "strconv"

import . "don/core"

func Par(coms []Com) (pc ParCom) {
	pc.Inners = make(map[int]Com)
	pc.InputType = NullType
	pc.OutputType = NullType
	for i, com := range coms {
		if _, nullp := com.(NullCom); !nullp {
			pc.Inners[i] = com
			pc.InputType.Joins(com.InputType())
			pc.OutputType.Joins(com.OutputType())
		}
	}
	pc.meetTypes()
	return
}

type ParCom struct {
	Inners                map[int]Com
	InputType, OutputType DType
}

// Mutates
func (pc *ParCom) meetTypes() {
	for {
		inputJoin, outputJoin := NullType, NullType
		for i, inner := range pc.Inners {
			if !inner.InputType().LTE(pc.InputType) || !inner.OutputType().LTE(pc.OutputType) {
				inner = inner.MeetTypes(pc.InputType, pc.OutputType)
				if _, nullp := inner.(NullCom); nullp {
					delete(pc.Inners, i)
				} else {
					pc.Inners[i] = inner
				}
			}
			inputJoin.Joins(inner.InputType())
			outputJoin.Joins(inner.OutputType())
		}
		if pc.InputType.LTE(inputJoin) && pc.OutputType.LTE(outputJoin) {
			break
		}
		pc.InputType.Meets(inputJoin)
		pc.OutputType.Meets(outputJoin)
	}
}

// Mutates
func (pc *ParCom) MeetTypes(inputType, outputType DType) {
	pc.InputType.Meets(inputType)
	pc.OutputType.Meets(outputType)
	pc.meetTypes()
}

func (pc ParCom) Underdefined(parName string) (underdefined Error) {
	for i, inner := range pc.Inners {
		underdefined.Ors(
			inner.Underdefined().Context("in " + strconv.Itoa(i) + "'th computer in " + parName),
		)
	}
	return
}

func (pc ParCom) Copy() ParCom {
	inners := make(map[int]Com, len(pc.Inners))
	for i, inner := range pc.Inners {
		inners[i] = inner.Copy()
	}
	pc.Inners = inners
	return pc
}

// Mutates
func (pc *ParCom) Invert() {
	for i, inner := range pc.Inners {
		pc.Inners[i] = inner.Invert()
	}
	pc.InputType, pc.OutputType = pc.OutputType, pc.InputType
}

func foreachWith(one TypeMap, many map[int]TypeMap, fn func(one Var, many []Var)) {
	for fieldName, subOne := range one.Fields {
		subMany := make(map[int]TypeMap)
		for i, manyElem := range many {
			if manyElemField, ok := manyElem.Fields[fieldName]; ok {
				subMany[i] = manyElemField
			}
		}
		foreachWith(subOne, subMany, fn)
	}
	if one.Unit != nil {
		var manyVars []Var
		for _, manyElem := range many {
			if manyElem.Unit != nil {
				manyVars = append(manyVars, manyElem.Unit)
			}
		}
		fn(one.Unit, manyVars)
	}
}

func (pc ParCom) TypedCom(
	tcb TypedComBuilder, /* mutated */
	inputMap, outputMap TypeMap,
	inputFn, outputFn func(Var, []Var),
) {
	innerInputMaps := make(map[int]TypeMap, len(pc.Inners))
	innerOutputMaps := make(map[int]TypeMap, len(pc.Inners))
	for i, inner := range pc.Inners {
		innerInputMaps[i] = MakeTypeMap(inner.InputType())
		innerOutputMaps[i] = MakeTypeMap(inner.OutputType())
		inner.TypedCom(tcb, innerInputMaps[i], innerOutputMaps[i])
	}
	foreachWith(inputMap, innerInputMaps, inputFn)
	foreachWith(outputMap, innerOutputMaps, outputFn)
}
