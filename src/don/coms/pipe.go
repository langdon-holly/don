package coms

import . "don/core"

/* len(coms) > 0 */
func Pipe(coms []Com) Com {
	comEntries := make([]CompositeComEntry, len(coms))
	for i := 0; i < len(coms)-1; i++ {
		comEntries[i] =
			CompositeComEntry{
				Com: coms[i],
				OutputMap: SignalReaderIdTree{
					SignalReaderId: SignalReaderId{
						ReaderId: ReaderId{InternalP: true, InternalIdx: i + 1}}}}
	}
	comEntries[len(coms)-1].Com = coms[len(coms)-1]

	inputMap := SignalReaderIdTree{
		SignalReaderId: SignalReaderId{
			ReaderId: ReaderId{InternalP: true}}}

	return CompositeCom{Coms: comEntries, InputMap: inputMap}
}
