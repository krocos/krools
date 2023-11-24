package krools

import (
	"context"
)

type (
	Action[T any] interface {
		Execute(ctx context.Context, fact T) error
	}
	Satisfiable[T any] interface {
		IsSatisfiedBy(ctx context.Context, candidate T) (bool, error)
	}
)

type ActionFn[T any] func(ctx context.Context, fact T) error

func (f ActionFn[T]) Execute(ctx context.Context, fact T) error { return f(ctx, fact) }

type ConditionFn[T any] func(ctx context.Context, candidate T) (bool, error)

func (f ConditionFn[T]) IsSatisfiedBy(ctx context.Context, candidate T) (bool, error) {
	return f(ctx, candidate)
}
