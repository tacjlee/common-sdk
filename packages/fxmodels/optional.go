package fxmodels

type Optional[T any] struct {
	Value *T
}

func (o *Optional[T]) IsPresent() bool {
	return o.Value != nil
}

func (o *Optional[T]) Get() *T {
	// If the caller does not need to modify the returned Value
	// If the caller needs to modify the returned Value, return a pointer.
	return o.Value
}

func (o *Optional[T]) IsNil() bool {
	return o.Value == nil
}

func (o *Optional[T]) GetPresentOrEmpty() *T {
	var model T
	if o.Value != nil {
		return o.Value
	} else {
		return &model
	}
}
