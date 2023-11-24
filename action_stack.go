package krools

import (
	"context"
)

type ActionStack[T any] struct {
	actions []Action[T]
}

func NewActionStack[T any](actions ...Action[T]) *ActionStack[T] {
	return &ActionStack[T]{actions: actions}
}

func (s *ActionStack[T]) Execute(ctx context.Context, fact T) error {
	for _, action := range s.actions {
		if err := action.Execute(ctx, fact); err != nil {
			return err
		}
	}

	return nil
}
