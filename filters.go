package krools

import (
	"context"
	"regexp"
	"strings"
)

type Filter interface {
	IsSatisfiedBy(ctx context.Context, rule *RuleHandle) (bool, error)
}

type FilterFn func(ctx context.Context, rule *RuleHandle) (bool, error)

func (f FilterFn) IsSatisfiedBy(ctx context.Context, rule *RuleHandle) (bool, error) {
	return f(ctx, rule)
}

func RuleNameStartsWith(prefix string) Filter {
	return FilterFn(func(ctx context.Context, rule *RuleHandle) (bool, error) {
		return strings.HasPrefix(rule.name, prefix), nil
	})
}

func RuleNameEndsWith(suffix string) Filter {
	return FilterFn(func(ctx context.Context, rule *RuleHandle) (bool, error) {
		return strings.HasSuffix(rule.name, suffix), nil
	})
}

func RuleNameMatchRegexp(exp *regexp.Regexp) Filter {
	return FilterFn(func(ctx context.Context, rule *RuleHandle) (bool, error) {
		return exp.MatchString(rule.name), nil
	})
}

func RuleNameMustNotContainsAny(substrings ...string) Filter {
	return FilterFn(func(ctx context.Context, rule *RuleHandle) (bool, error) {
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

func RuleNameMustContainsAny(substrings ...string) Filter {
	return FilterFn(func(ctx context.Context, rule *RuleHandle) (bool, error) {
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

func RunOnlyUnits(units ...string) Filter {
	return FilterFn(func(ctx context.Context, rule *RuleHandle) (bool, error) {
		if len(units) > 0 {
			return contains(units, rule.unit), nil
		}

		return true, nil
	})
}
