package coms

import . "don/core"

// len(coms) > 0
func GenPipe(genComs []GenCom) GenCom {
	genComEntries := make([]GenCompositeEntry, len(genComs))
	for i := 0; i < len(genComs)-1; i++ {
		genComEntries[i] =
			GenCompositeEntry{
				GenCom: genComs[i],
				OutputMap: SignalReaderIdTree{
					SignalReaderId: SignalReaderId{
						ReaderId: ReaderId{InternalP: true, InternalIdx: i + 1}}}}
	}
	genComEntries[len(genComs)-1].GenCom = genComs[len(genComs)-1]

	inputMap := SignalReaderIdTree{
		SignalReaderId: SignalReaderId{
			ReaderId: ReaderId{InternalP: true}}}

	return GenComposite{GenComs: genComEntries, InputMap: inputMap}
}
