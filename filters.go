package krools

import (
	"context"
	"strings"
)

func RuleNameStartsWith[T any](prefix string) Satisfiable[*Rule[T]] {
	return ConditionFn[*Rule[T]](func(ctx context.Context, rule *Rule[T]) (bool, error) {
		return strings.HasPrefix(rule.name, prefix), nil
	})
}

func RuleNameEndsWith[T any](suffix string) Satisfiable[*Rule[T]] {
	return ConditionFn[*Rule[T]](func(ctx context.Context, rule *Rule[T]) (bool, error) {
		return strings.HasSuffix(rule.name, suffix), nil
	})
}

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
	return ConditionFn[*Rule[T]](func(ctx context.Context, rule *Rule[T]) (bool, error) {
		if len(units) > 0 {
			return contains(units, rule.unit), nil
		}

		return true, nil
	})
}
