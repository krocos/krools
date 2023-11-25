package krools

import (
	"context"
	"errors"
	"fmt"
)

type Set[T any] struct {
	name              string
	agendaGroups      map[string][]*Rule[T]
	agendaGroupsOrder []string
	activationGroups  map[string][]*Rule[T]
	maxReevaluations  int
}

func NewSet[T any](name string) *Set[T] {
	return &Set[T]{
		name:             name,
		agendaGroups:     make(map[string][]*Rule[T]),
		activationGroups: make(map[string][]*Rule[T]),
		maxReevaluations: 256,
	}
}

func (s *Set[T]) Add(rule *Rule[T]) *Set[T] {
	var agendaGroupRules []*Rule[T]

	for _, existing := range s.agendaGroups[rule.agendaGroup] {
		if existing.name != rule.name {
			agendaGroupRules = append(agendaGroupRules, existing)
		}
	}

	s.agendaGroups[rule.agendaGroup] = append(agendaGroupRules, rule)
	s.agendaGroupsOrder = uniq(append(s.agendaGroupsOrder, rule.agendaGroup))

	if rule.activationGroup != nil {
		var activationGroupRules []*Rule[T]

		for _, existing := range s.activationGroups[*rule.activationGroup] {
			if existing.name != rule.name {
				activationGroupRules = append(activationGroupRules, existing)
			}
		}

		s.activationGroups[*rule.activationGroup] = append(activationGroupRules, rule)
	}

	return s
}

func (s *Set[T]) SetFocus(agendaGroups ...string) *Set[T] {
	s.agendaGroupsOrder = uniq(append(agendaGroups, s.agendaGroupsOrder...))

	return s
}

func (s *Set[T]) SetMaxReevaluations(v int) *Set[T] {
	s.maxReevaluations = v

	return s
}

func (s *Set[T]) FireAllRules(ctx context.Context, fact T) error {
	ret := newRetracting()

	var reevaluations int

	for _, agendaGroup := range s.agendaGroupsOrder {
		applicable, err := s.applicableRules(ctx, agendaGroup, fact, ret)
		if err != nil {
			return err
		}

		for len(applicable) > 0 {
			for _, rule := range applicable {
				if err = s.executeAction(ctx, fact, rule, ret); err != nil {
					return err
				}
			}

			applicable, err = s.applicableRules(ctx, agendaGroup, fact, ret)
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

func (s *Set[T]) applicableRules(ctx context.Context, agendaGroup string, fact T, ret *retracting) ([]*Rule[T], error) {
	var applicable []*Rule[T]

	for _, rule := range s.agendaGroups[agendaGroup] {
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

	if rule.activationGroup != nil {
		var names []string

		for _, r := range s.activationGroups[*rule.activationGroup] {
			names = append(names, r.name)
		}

		ret.add(reject(names, rule.name)...)
	}

	return nil
}
