package fxoperators

func If[T any](condition bool, trueVal interface{}, falseVal interface{}) T {
	if condition {
		return trueVal.(T)
	}
	return falseVal.(T)
}
