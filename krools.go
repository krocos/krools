package krools

import (
	"cmp"
	"context"
	"slices"
)

const UnitMAIN = "MAIN"

type Context interface {
	Context() context.Context

	Set(v any)
	Get(v any) bool
	Handle(v any) any
	HasNot(v any) bool
	Delete(v any)

	SetLocal(v any)
	GetLocal(v any) bool
	LocalHandle(v any) any
	HasNotLocal(v any) bool
	DeleteLocal(v any)
}

type ActionFn func(ctx Context) error

func (f ActionFn) Then(ctx Context) error { return f(ctx) }

type ConditionFn func(ctx Context) (bool, error)

func (f ConditionFn) When(ctx Context) (bool, error) { return f(ctx) }

func sortRulesConsiderSalience(rules []*RuleHandle) {
	slices.SortFunc(rules, func(a, b *RuleHandle) int {
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
