package coms

import . "don/core"

func makeMaps(map0, map1 *interface{}, chanN *CompositeComChanSourceN, theType DType) {
	switch theType.Tag {
	case UnitTypeTag:
		*map0 = chanN.Units
		*map1 = chanN.Units
		chanN.Units++
	case SyntaxTypeTag:
		*map0 = chanN.Syntaxen
		*map1 = chanN.Syntaxen
		chanN.Syntaxen++
	case GenComTypeTag:
		*map0 = chanN.GenComs
		*map1 = chanN.GenComs
		chanN.GenComs++
	case StructTypeTag:
		map0Val := make(Struct)
		*map0 = map0Val

		map1Val := make(Struct)
		*map1 = map1Val

		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			var fieldMap0 interface{}
			var fieldMap1 interface{}

			makeMaps(&fieldMap0, &fieldMap1, chanN, fieldType)

			map0Val[fieldName] = fieldMap0
			map1Val[fieldName] = fieldMap1
		}
	}
}

// len(coms) > 0
func pipe(coms []Com) (ret CompositeCom) {
	ret.TheInputType = coms[0].InputType()
	ret.TheOutputType = coms[len(coms)-1].OutputType()

	ret.ComEntries = make([]CompositeComEntry, len(coms))
	for i, com := range coms {
		ret.ComEntries[i].Com = com
	}

	for i := 0; i < len(ret.ComEntries)-1; i++ {
		//TODO: Check types
		makeMaps(&ret.ComEntries[i].OutputMap, &ret.ComEntries[i+1].InputMap, &ret.InputChanN, ret.ComEntries[i].OutputType())
	}

	ret.OutputChanN = ret.InputChanN
	ret.InnerChanN = ret.InputChanN

	makeMaps(&ret.InputMap, &ret.ComEntries[0].InputMap, &ret.InputChanN, ret.TheInputType)
	makeMaps(&ret.OutputMap, &ret.ComEntries[len(coms)-1].OutputMap, &ret.OutputChanN, ret.TheOutputType)

	return
}

func GenPipe(genComs []GenCom) GenCom {
	return func(inputType DType) Com {
		coms := make([]Com, len(genComs))
		for i, genCom := range genComs {
			com := genCom(inputType)
			coms[i] = com
			inputType = com.OutputType()
		}
		return pipe(coms)
	}
}
