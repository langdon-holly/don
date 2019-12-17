package types

import . "don/core"

var SyntaxType = MakeStructType(make(map[string]DType, 6))

func init() {
	_, syntaxen1Type := MakeLinkedListType(SyntaxType)
	SyntaxType.Fields["block"], _ = MakeLinkedListType(syntaxen1Type) /* for block? */

	mCallFields := make(map[string]DType, 2)
	mCallFields["macro"] = SyntaxType
	mCallFields["param"] = SyntaxType
	SyntaxType.Fields["macro-call"] = MakeRefType(MakeStructType(mCallFields))

	SyntaxType.Fields["ident?"] = UnitType
	SyntaxType.Fields["select?"] = UnitType
	SyntaxType.Fields["deselect?"] = UnitType

	_, SyntaxType.Fields["field-path"] = MakeLinkedListType(BytesType) /* for ident? or select? or deselect? */
}
