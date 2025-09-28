package fxoperator

func If[T any](condition bool, trueVal any, falseVal any) T {
	if condition {
		return trueVal.(T)
	}
	return falseVal.(T)
}
