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

func RuleNameMustNotContainsAny[T any](substrings ...string) Satisfiable[*Rule[T]] {
	return ConditionFn[*Rule[T]](func(ctx context.Context, rule *Rule[T]) (bool, error) {
		if len(substrings) > 0 {
			for _, substring := range substrings {
				if strings.Contains(rule.name, substring) {
					return false, nil
				}
			}
		}

		return true, nil
	})
}

func RuleNameMustContainsAny[T any](substrings ...string) Satisfiable[*Rule[T]] {
	return ConditionFn[*Rule[T]](func(ctx context.Context, rule *Rule[T]) (bool, error) {
		if len(substrings) > 0 {
			for _, substring := range substrings {
				if strings.Contains(rule.name, substring) {
					return true, nil
				}
			}

			return false, nil
		}

		return true, nil
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
