package coms

import . "don/core"

func Deselect(fieldName string) Com {
	return CompositeCom{
		Coms: []CompositeComEntry{{
			Com: ICom{},
			OutputMap: SignalMap{
				ExternalP: true,
				FieldPath: []string{fieldName}}}},
		InputMap: SignalMap{InternalIdx: 0}}
}
