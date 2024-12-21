package fxarraylist

import (
	"errors"
)

// ArrayList is a generic list implementation
type ArrayList[T any] struct {
	elements []T
}

// NewArrayList creates a new ArrayList
func NewArrayList[T any]() *ArrayList[T] {
	return &ArrayList[T]{elements: []T{}}
}

// Add appends an element to the ArrayList
func (list *ArrayList[T]) Add(element T) {
	list.elements = append(list.elements, element)
}

// Get retrieves the element at the specified index
func (list *ArrayList[T]) Get(index int) (T, error) {
	if index < 0 || index >= len(list.elements) {
		var zeroValue T
		return zeroValue, errors.New("index out of bounds")
	}
	return list.elements[index], nil
}

// Set replaces the element at the specified index with the given element
func (list *ArrayList[T]) Set(index int, element T) error {
	if index < 0 || index >= len(list.elements) {
		return errors.New("index out of bounds")
	}
	list.elements[index] = element
	return nil
}

// Remove removes the element at the specified index
func (list *ArrayList[T]) Remove(index int) (T, error) {
	if index < 0 || index >= len(list.elements) {
		var zeroValue T
		return zeroValue, errors.New("index out of bounds")
	}
	removed := list.elements[index]
	list.elements = append(list.elements[:index], list.elements[index+1:]...)
	return removed, nil
}

// Size returns the number of elements in the ArrayList
func (list *ArrayList[T]) Size() int {
	return len(list.elements)
}

// Clear removes all elements from the ArrayList
func (list *ArrayList[T]) Clear() {
	list.elements = []T{}
}
