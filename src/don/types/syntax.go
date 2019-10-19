package types

import . "don/core"

var SyntaxType = MakeStructType(make(map[string]DType, 3))

func init() {
	SyntaxType.Fields["block"] = MakeLinkedListType(*MakeLinkedListType(SyntaxType).Referent)

	mCallFields := make(map[string]DType, 2)
	mCallFields["macro"] = SyntaxType
	mCallFields["param"] = SyntaxType
	SyntaxType.Fields["mcall"] = MakeRefType(MakeStructType(mCallFields))

	SyntaxType.Fields["ident"] = BytesType
}
