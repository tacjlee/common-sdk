package fxmodels

type Optional[T any] struct {
	Value *T
}

// Chỉ đọc dữ liệu, bạn có thể dùng Optional[T] thay vì *Optional[T]
// o *Optional[T] là con trỏ, thao tác trực tiếp trên đối tượng gốc.
// o Optional[T] là giá trị, thao tác trên một bản sao.
func (o *Optional[T]) IsPresent() bool {
	return o.Value != nil
}

func (o *Optional[T]) Get() T {
	// If the caller does not need to modify the returned Value
	if o.Value == nil {
		var zero T
		return zero
	}
	return *o.Value
}

func (o *Optional[T]) GetPointer() *T {
	// If the caller needs to modify the returned Value, return a pointer.
	return o.Value
}

func (o *Optional[T]) GetPresentOrEmpty() T {
	if o.Value != nil {
		return *o.Value
	}
	var model T
	return model
}
