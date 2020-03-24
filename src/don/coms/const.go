package coms

import . "don/core"

type ConstVal struct {
	P         bool /* for !struct */
	RefVal    Ref
	StructVal map[string]ConstVal
}

func Const(outputType DType, val ConstVal) Com {
	switch outputType.Tag {
	case UnitTypeTag:
		if val.P {
			return ICom{}
		} else {
			return Sink
		}
	case RefTypeTag:
		if val.P {
			return ConstRefCom{ReferentType: *outputType.Referent, Val: val.RefVal}
		} else {
			return Sink
		}
	case StructTypeTag:
		inputMap := SignalMap{InternalIdx: 0}

		fieldNames := make([]string, len(outputType.Fields))
		splitOutputMapFields := make(map[string]SignalMap, len(outputType.Fields))
		coms := make([]CompositeComEntry, len(outputType.Fields)+1)

		i := 0
		for fieldName, fieldType := range outputType.Fields {
			splitOutputMapFields[fieldName] = SignalMap{InternalIdx: i}
			coms[i+1] = CompositeComEntry{
				Com: Const(fieldType, val.StructVal[fieldName]),
				OutputMap: SignalMap{
					ExternalP: true,
					FieldPath: []string{fieldName}}}
			fieldNames[i] = fieldName
			i++
		}

		coms[0] = CompositeComEntry{
			Com: SplitCom(fieldNames),
			OutputMap: SignalMap{
				ParentP:  true,
				Children: splitOutputMapFields}}

		return CompositeCom{Coms: coms, InputMap: inputMap}
	default:
		panic("Unreachable")
	}
}
