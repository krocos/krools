package krools

func NewStatefulRule[T any](state func() *Rule[T]) *Rule[T] {
	return state()
}
