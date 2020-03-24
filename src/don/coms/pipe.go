package coms

import . "don/core"

/* len(coms) > 0 */
func Pipe(coms []Com) Com {
	comEntries := make([]CompositeComEntry, len(coms))
	for i := 0; i < len(coms)-1; i++ {
		comEntries[i] = CompositeComEntry{
			Com:       coms[i],
			OutputMap: SignalMap{InternalIdx: i + 1}}
	}
	comEntries[len(coms)-1] = CompositeComEntry{
		Com:       coms[len(coms)-1],
		OutputMap: SignalMap{ExternalP: true}}

	inputMap := SignalMap{InternalIdx: 0}

	return CompositeCom{Coms: comEntries, InputMap: inputMap}
}
