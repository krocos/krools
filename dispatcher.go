package krools

import (
	"context"
)

type Executor[T any] func(ctx context.Context, fireContext T) error

type Dispatcher[T any] interface {
	Dispatch(ctx context.Context, fireContext T, fireAllRules Executor[T]) error
}
