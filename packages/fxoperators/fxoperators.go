package fxoperators

func If(condition bool, trueVal any, falseVal any) any {
	if condition {
		return trueVal
	}
	return falseVal
}
