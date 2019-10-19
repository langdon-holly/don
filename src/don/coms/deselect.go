package coms

import . "don/core"

func Deselect(fieldName string) Com {
	return CompositeCom{
		Coms: []CompositeComEntry{{
			Com: ICom{},
			OutputMap: SignalReaderIdTree{
				SignalReaderId: SignalReaderId{FieldPath: []string{fieldName}}}}},
		InputMap: SignalReaderIdTree{
			SignalReaderId: SignalReaderId{ReaderId: ReaderId{InternalP: true}}}}
}
