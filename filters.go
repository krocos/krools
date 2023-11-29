package krools

import (
	"context"
	"strings"
)

func RuleNameMustNotContains[T any](s string) Satisfiable[*Rule[T]] {
	return ConditionFn[*Rule[T]](func(ctx context.Context, rule *Rule[T]) (bool, error) {
		return !strings.Contains(rule.name, s), nil
	})
}

func RuleNameMustContains[T any](s string) Satisfiable[*Rule[T]] {
	return ConditionFn[*Rule[T]](func(ctx context.Context, rule *Rule[T]) (bool, error) {
		return strings.Contains(rule.name, s), nil
	})
}

func RunOnlyRulesFromUnits[T any](units ...string) Satisfiable[*Rule[T]] {
	return ConditionFn[*Rule[T]](func(ctx context.Context, candidate *Rule[T]) (bool, error) {
		if len(units) > 0 {
			return contains(units, candidate.unit), nil
		}

		return true, nil
	})
}
