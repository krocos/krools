package krools

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type KnowledgeBase[T any] struct {
	name             string
	units            map[string][]*Rule[T]
	unitsOrder       []string
	activationUnits  map[string][]*Rule[T]
	deactivatedUnits []string
	maxReevaluations int
}

func NewKnowledge[T any](name string) *KnowledgeBase[T] {
	return &KnowledgeBase[T]{
		name:             name,
		units:            make(map[string][]*Rule[T]),
		activationUnits:  make(map[string][]*Rule[T]),
		maxReevaluations: 256,
	}
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

func (s *KnowledgeBase[T]) Add(rule *Rule[T]) *KnowledgeBase[T] {
	var units []*Rule[T]

	for _, existing := range s.units[rule.unit] {
		if existing.name != rule.name {
			units = append(units, existing)
		}
	}

	s.units[rule.unit] = append(units, rule)
	s.unitsOrder = uniq(append(s.unitsOrder, rule.unit))

	if rule.activationUnit != nil {
		var activationUnits []*Rule[T]

		for _, existing := range s.activationUnits[*rule.activationUnit] {
			if existing.name != rule.name {
				activationUnits = append(activationUnits, existing)
			}
		}

		s.activationUnits[*rule.activationUnit] = append(activationUnits, rule)
	}

	return s
}

func (s *KnowledgeBase[T]) SetFocus(units ...string) *KnowledgeBase[T] {
	s.unitsOrder = uniq(append(units, s.unitsOrder...))

	return s
}

func (s *KnowledgeBase[T]) SetDeactivatedUnits(units ...string) *KnowledgeBase[T] {
	s.deactivatedUnits = uniq(append(units, s.deactivatedUnits...))

	return s
}

func (s *KnowledgeBase[T]) SetMaxReevaluations(v int) *KnowledgeBase[T] {
	s.maxReevaluations = v

	return s
}

func (s *KnowledgeBase[T]) FireAllRules(ctx context.Context, fact T, ruleFilters ...Satisfiable[*Rule[T]]) error {
	ret := newRetracting()
	flow := newFlowController[T](ret, s.units, s.unitsOrder, s.deactivatedUnits)

	var reevaluations int

	for flow.more() {
		applicable, err := s.applicableRules(ctx, flow.rules(), fact, ret, ruleFilters...)
		if err != nil {
			return err
		}

		for len(applicable) > 0 {
			for _, rule := range applicable {
				if err = s.executeAction(ctx, fact, rule, ret, flow); err != nil {
					return err
				}
			}

			applicable, err = s.applicableRules(ctx, flow.rules(), fact, ret, ruleFilters...)
			if err != nil {
				return err
			}

			reevaluations++
			if reevaluations > s.maxReevaluations {
				return errors.New("too much reevaluations")
			}
		}
	}

	return nil
}

func (s *KnowledgeBase[T]) applicableRules(
	ctx context.Context,
	rules []*Rule[T],
	fact T,
	ret *retracting,
	filters ...Satisfiable[*Rule[T]],
) ([]*Rule[T], error) {
	var applicable []*Rule[T]

loop:
	for _, rule := range rules {
		if ret.isRetracted(rule.name) {
			continue
		}

		for i, filter := range filters {
			ok, err := filter.IsSatisfiedBy(ctx, rule)
			if err != nil {
				return nil, fmt.Errorf("verify that rule '%s' of knowledge base '%s' is satisfies filter %d: %w", rule.name, s.name, i, err)
			}

			if !ok {
				continue loop
			}
		}

		satisfied := true

		if rule.condition != nil {
			var err error
			satisfied, err = rule.condition.IsSatisfiedBy(ctx, fact)
			if err != nil {
				return nil, fmt.Errorf("verify that condition of rule '%s' of knowledge base '%s' is satisfied by fact: %w", rule.name, s.name, err)
			}
		}

		if satisfied {
			applicable = append(applicable, rule)
		}
	}

	sortRulesConsiderSalience(applicable)

	return applicable, nil
}

func (s *KnowledgeBase[T]) executeAction(ctx context.Context, fact T, rule *Rule[T], ret *retracting, flow *flowController[T]) error {
	if ret.isRetracted(rule.name) {
		return nil
	}

	if err := rule.action.Execute(ctx, fact); err != nil {
		return fmt.Errorf("execute action of rule '%s' of knowledge base '%s': %w", rule.name, s.name, err)
	}

	ret.add(rule.retracts...)
	flow.deactivateUnits(rule.deactivateUnits...)
	flow.activateUnits(rule.activateUnits...)
	flow.setFocus(rule.focusUnits...)

	if rule.activationUnit != nil {
		var names []string

		for _, r := range s.activationUnits[*rule.activationUnit] {
			names = append(names, r.name)
		}

		ret.add(reject(names, rule.name)...)
	}

	return nil
}