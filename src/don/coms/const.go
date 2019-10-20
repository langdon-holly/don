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
		inputMap := SignalReaderIdTree{
			SignalReaderId: SignalReaderId{
				ReaderId: ReaderId{InternalP: true}}}

		fieldNames := make([]string, len(outputType.Fields))
		splitOutputMapFields := make(map[string]SignalReaderIdTree, len(outputType.Fields))
		coms := make([]CompositeComEntry, len(outputType.Fields)+1)

		i := 0
		for fieldName, fieldType := range outputType.Fields {
			splitOutputMapFields[fieldName] = SignalReaderIdTree{
				SignalReaderId: SignalReaderId{
					ReaderId: ReaderId{InternalP: true, InternalIdx: i}}}
			coms[i+1] = CompositeComEntry{
				Com: Const(fieldType, val.StructVal[fieldName]),
				OutputMap: SignalReaderIdTree{
					SignalReaderId: SignalReaderId{
						FieldPath: []string{fieldName}}}}
			fieldNames[i] = fieldName
			i++
		}

		coms[0] = CompositeComEntry{Com: SplitCom(fieldNames), OutputMap: SignalReaderIdTree{ParentP: true, Children: splitOutputMapFields}}

		return CompositeCom{Coms: coms, InputMap: inputMap}
	default:
		panic("Unreachable")
	}
}
