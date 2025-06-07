package utils

import "fmt"

type StackUnderflow struct {}

func (s StackUnderflow) Error() string {
	return "Stack underflow"
}

type Stack[T any] struct {
	Size int
	data []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		Size: 0,
		data: []T{},
	}
}

func (s *Stack[T]) Push(value T) {
	s.data = append(s.data, value)
	s.Size++
}

func (s *Stack[T]) Pop() (T, error) {
	if s.Size == 0 {
		var e T
		return e, &StackUnderflow{}
	}
	top := s.data[len(s.data) - 1]
	s.data = s.data[:len(s.data) - 1]
	s.Size--
	return top, nil
}

func (s Stack[T]) Top() (T, error) {
	if s.Size == 0 {
		var e T
		return e, &StackUnderflow{}
	}

	return s.data[len(s.data) - 1], nil
}	

func (s *Stack[T]) Data() []T {
	copied := make([]T, len(s.data))
	copy(copied, s.data)
	return copied
}

func (s Stack[T]) String() string {
	result := ""
	
	for _, element := range s.data {
		result += fmt.Sprintf("[ %v ]\n", element)	
	}

	return result
}
