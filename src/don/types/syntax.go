package types

import . "don/core"

var SyntaxType = MakeStructType(make(map[string]DType, 3))

func init() {
	SyntaxType.Fields["block?"] = UnitType
	SyntaxType.Fields["block"] = MakeLinkedListType(*MakeLinkedListType(SyntaxType).Referent)

	mCallFields := make(map[string]DType, 2)
	mCallFields["macro"] = SyntaxType
	mCallFields["param"] = SyntaxType
	SyntaxType.Fields["macro-call"] = MakeRefType(MakeStructType(mCallFields))

	SyntaxType.Fields["ident?"] = UnitType
	SyntaxType.Fields["select?"] = UnitType
	SyntaxType.Fields["deselect?"] = UnitType
	SyntaxType.Fields["field-path"] = *MakeLinkedListType(BytesType).Referent
}
