package coms

import "strconv"

import . "don/core"

// len(pipes) >= 1
func ComCom(pipes []Com) Com {
	if len(pipes) == 1 {
		return pipes[0]
	}

	// len(pipes) >= 2

	indexStrings := make([]string, len(pipes)-1)
	comEntries := make([]CompositeComEntry, len(pipes)+1)

	comEntries[len(pipes)].OutputMap = SignalReaderIdTree{
		ParentP:  true,
		Children: make(map[string]SignalReaderIdTree, len(pipes)-1)}

	{
		i := 0

		{
			indexString := strconv.FormatInt(int64(i), 10)
			indexStrings[i] = indexString

			comEntries[i].Com = pipes[i]
			comEntries[len(pipes)].OutputMap.Children[indexString] =
				SignalReaderIdTree{SignalReaderId: SignalReaderId{
					ReaderId: ReaderId{InternalP: true, InternalIdx: i}}}
			i++
		}

		for ; i < len(pipes)-1; i++ {
			indexString := strconv.FormatInt(int64(i), 10)
			indexStrings[i] = indexString

			comEntries[i] =
				CompositeComEntry{
					Com: pipes[i],
					OutputMap: SignalReaderIdTree{
						SignalReaderId: SignalReaderId{
							ReaderId:  ReaderId{InternalP: true, InternalIdx: len(pipes)},
							FieldPath: []string{indexString}}}}
			comEntries[len(pipes)].OutputMap.Children[indexString] =
				SignalReaderIdTree{SignalReaderId: SignalReaderId{
					ReaderId: ReaderId{InternalP: true, InternalIdx: i}}}
		}

		indexString := strconv.FormatInt(int64(i), 10)

		comEntries[i] =
			CompositeComEntry{
				Com: pipes[i],
				OutputMap: SignalReaderIdTree{
					SignalReaderId: SignalReaderId{
						ReaderId:  ReaderId{InternalP: true, InternalIdx: len(pipes)},
						FieldPath: []string{indexString}}}}
	}

	comEntries[len(pipes)].Com = Pipe([]Com{MergeCom{}, SplitCom(indexStrings)})

	inputMap := SignalReaderIdTree{
		SignalReaderId: SignalReaderId{
			ReaderId: ReaderId{InternalP: true, InternalIdx: len(pipes) - 1}}}

	return CompositeCom{Coms: comEntries, InputMap: inputMap}
}
