package utils

import "fmt"

type Optional[T any] struct {
	value  T
	hasVal bool
}

func (o Optional[T]) String() string {
	return fmt.Sprintf("[%v %v]", o.value, o.hasVal)
}

func (o *Optional[T]) HasVal() bool {
	return o.hasVal
}

func (o *Optional[T]) Value() T {
	return o.value
}

func (o *Optional[T]) SetVal(value T) {
	o.value = value
	o.hasVal = true
}

func Some[T any](value T) *Optional[T] {
	return &Optional[T]{
		value: value,
		hasVal: true,
	}
}

func None[T any]() *Optional[T] {
	return &Optional[T]{
		hasVal: false,
	}
}
