package coms

import "strconv"

import . "don/core"

func SplitMerge(coms []Com) Com {
	indexStrings := make([]string, len(coms))
	comEntries := make([]CompositeComEntry, len(coms)+2)

	comEntries[len(coms)].OutputMap = SignalReaderIdTree{
		ParentP:  true,
		Children: make(map[string]SignalReaderIdTree, len(coms))}

	for i := 0; i < len(coms); i++ {
		indexString := strconv.FormatInt(int64(i), 10)
		indexStrings[i] = indexString

		comEntries[i] =
			CompositeComEntry{
				Com: coms[i],
				OutputMap: SignalReaderIdTree{
					SignalReaderId: SignalReaderId{
						ReaderId:  ReaderId{InternalP: true, InternalIdx: len(coms) + 1},
						FieldPath: []string{indexString}}}}
		comEntries[len(coms)].OutputMap.Children[indexString] =
			SignalReaderIdTree{SignalReaderId: SignalReaderId{
				ReaderId: ReaderId{InternalP: true, InternalIdx: i}}}
	}
	comEntries[len(coms)].Com = SplitCom(indexStrings)
	comEntries[len(coms)+1].Com = MergeCom{}

	inputMap := SignalReaderIdTree{
		SignalReaderId: SignalReaderId{
			ReaderId: ReaderId{InternalP: true, InternalIdx: len(coms)}}}

	return CompositeCom{Coms: comEntries, InputMap: inputMap}
}
