package krools

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
)

type Set[T any] struct {
	name             string
	rules            map[string]*Rule[T]
	maxReevaluations int
}

func NewSet[T any](name string) *Set[T] {
	return &Set[T]{
		name:             name,
		rules:            make(map[string]*Rule[T]),
		maxReevaluations: 256,
	}
}

func (s *Set[T]) Add(rule *Rule[T]) *Set[T] {
	s.rules[rule.name] = rule

	return s
}

func (s *Set[T]) SetMaxReevaluations(v int) *Set[T] {
	s.maxReevaluations = v

	return s
}

func (s *Set[T]) FireAllRules(ctx context.Context, fact T) error {
	ret := newRetracting()

	applicable, err := s.applicableRules(ctx, fact, ret)
	if err != nil {
		return err
	}

	var reevaluations int

	for len(applicable) > 0 {
		for _, rule := range applicable {
			if err = s.executeAction(ctx, fact, rule, ret); err != nil {
				return err
			}
		}

		applicable, err = s.applicableRules(ctx, fact, ret)
		if err != nil {
			return err
		}

		reevaluations++
		if reevaluations > s.maxReevaluations {
			return errors.New("too much reevaluations")
		}
	}

	return nil
}

func (s *Set[T]) applicableRules(ctx context.Context, fact T, ret *retracting) ([]*Rule[T], error) {
	var applicable []*Rule[T]

	for _, rule := range s.rules {
		if ret.isRetracted(rule.name) {
			continue
		}

		satisfied, err := rule.condition.IsSatisfiedBy(ctx, fact)
		if err != nil {
			return nil, fmt.Errorf("verify that condition of rule '%s' of set '%s' is satisfied by fact: %w", rule.name, s.name, err)
		}

		if satisfied {
			applicable = append(applicable, rule)
		}
	}

	sortRulesConsiderSalience(applicable)

	return applicable, nil
}

func (s *Set[T]) executeAction(ctx context.Context, fact T, rule *Rule[T], ret *retracting) error {
	if ret.isRetracted(rule.name) {
		return nil
	}

	if err := rule.action.Execute(ctx, fact); err != nil {
		return fmt.Errorf("execute action of rule '%s' of set '%s': %w", rule.name, s.name, err)
	}

	ret.add(rule.retracts...)

	return nil
}

func sortRulesConsiderSalience[T any](rules []*Rule[T]) {
	slices.SortFunc(rules, func(a, b *Rule[T]) int {
		return cmp.Compare(a.salience, b.salience) * -1
	})
}