package krools

import (
	"cmp"
	"context"
	"slices"
)

const UnitMAIN = "MAIN"

type (
	Action[T any] interface {
		Execute(ctx context.Context, fireContext T) error
	}
	Satisfiable[T any] interface {
		IsSatisfiedBy(ctx context.Context, fireContext T) (bool, error)
	}
)

type ActionFn[T any] func(ctx context.Context, fireContext T) error

func (f ActionFn[T]) Execute(ctx context.Context, fireContext T) error { return f(ctx, fireContext) }

type ConditionFn[T any] func(ctx context.Context, fireContext T) (bool, error)

func (f ConditionFn[T]) IsSatisfiedBy(ctx context.Context, fireContext T) (bool, error) {
	return f(ctx, fireContext)
}

func sortRulesConsiderSalience[T any](rules []*Rule[T]) {
	slices.SortFunc(rules, func(a, b *Rule[T]) int {
		return cmp.Compare(a.salience, b.salience) * -1
	})
}

func uniq[T comparable](collection []T) []T {
	result := make([]T, 0, len(collection))
	seen := make(map[T]struct{}, len(collection))

	for _, item := range collection {
		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		result = append(result, item)
	}

	return result
}

func reject(collection []string, value string) []string {
	result := make([]string, 0)

	for _, item := range collection {
		if item != value {
			result = append(result, item)
		}
	}

	return result
}

func contains[T comparable](collection []T, element T) bool {
	for _, item := range collection {
		if item == element {
			return true
		}
	}

	return false
}
