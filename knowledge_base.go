package krools

import (
	"context"
	"errors"
	"fmt"
)

type KnowledgeBase[T any] struct {
	name             string
	units            map[string][]*Rule[T]
	unitsOrder       []string
	activationUnits  map[string][]*Rule[T]
	deactivatedUnits []string
	maxReevaluations int
}

func NewKnowledgeBase[T any](name string) *KnowledgeBase[T] {
	return &KnowledgeBase[T]{
		name:             name,
		units:            make(map[string][]*Rule[T]),
		activationUnits:  make(map[string][]*Rule[T]),
		maxReevaluations: 65535,
	}
}

func (k *KnowledgeBase[T]) Add(rule *Rule[T]) *KnowledgeBase[T] {
	var units []*Rule[T]

	for _, existing := range k.units[rule.unit] {
		if existing.name != rule.name {
			units = append(units, existing)
		}
	}

	k.units[rule.unit] = append(units, rule)
	k.unitsOrder = uniq(append(k.unitsOrder, rule.unit))

	if rule.activationUnit != nil {
		var activationUnits []*Rule[T]

		for _, existing := range k.activationUnits[*rule.activationUnit] {
			if existing.name != rule.name {
				activationUnits = append(activationUnits, existing)
			}
		}

		k.activationUnits[*rule.activationUnit] = append(activationUnits, rule)
	}

	return k
}

func (k *KnowledgeBase[T]) AddUnit(unit string, rules ...*Rule[T]) *KnowledgeBase[T] {
	for _, rule := range rules {
		k.Add(rule.Unit(unit))
	}

	return k
}

func (k *KnowledgeBase[T]) SetFocus(units ...string) *KnowledgeBase[T] {
	k.unitsOrder = uniq(append(units, k.unitsOrder...))

	return k
}

func (k *KnowledgeBase[T]) SetDeactivatedUnits(units ...string) *KnowledgeBase[T] {
	k.deactivatedUnits = uniq(append(units, k.deactivatedUnits...))

	return k
}

func (k *KnowledgeBase[T]) SetMaxReevaluations(v int) *KnowledgeBase[T] {
	k.maxReevaluations = v

	return k
}

func (k *KnowledgeBase[T]) FireAllRules(ctx context.Context, fireContext T, ruleFilters ...Condition[*Rule[T]]) error {
	ret := newRetracting()
	flow := newFlowController[T](ret, k.units, k.unitsOrder, k.deactivatedUnits)

	var reevaluations int

	for flow.more() {
		applicable, err := k.applicableRules(ctx, flow.rules(), fireContext, ret, ruleFilters...)
		if err != nil {
			return err
		}

		for len(applicable) > 0 {
			for _, rule := range applicable {
				if err = k.executeAction(ctx, fireContext, rule, ret, flow); err != nil {
					return err
				}
			}

			applicable, err = k.applicableRules(ctx, flow.rules(), fireContext, ret, ruleFilters...)
			if err != nil {
				return err
			}

			reevaluations++
			if reevaluations > k.maxReevaluations {
				return errors.New("too much reevaluations")
			}
		}
	}

	return nil
}

func (k *KnowledgeBase[T]) applicableRules(
	ctx context.Context,
	rules []*Rule[T],
	fireContext T,
	ret *retracting,
	filters ...Condition[*Rule[T]],
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
				return nil, fmt.Errorf("verify that rule '%s' of knowledge base '%s' is satisfies filter %d: %w", rule.name, k.name, i, err)
			}

			if !ok {
				continue loop
			}
		}

		satisfied := true

		if rule.condition != nil {
			var err error
			satisfied, err = rule.condition.IsSatisfiedBy(ctx, fireContext)
			if err != nil {
				return nil, fmt.Errorf("verify that condition of rule '%s' of knowledge base '%s' is satisfied by fire context: %w", rule.name, k.name, err)
			}
		}

		if satisfied {
			applicable = append(applicable, rule)
		}
	}

	sortRulesConsiderSalience(applicable)

	return applicable, nil
}

func (k *KnowledgeBase[T]) executeAction(ctx context.Context, fireContext T, rule *Rule[T], ret *retracting, flow *flowController[T]) error {
	if ret.isRetracted(rule.name) {
		return nil
	}

	if rule.action != nil {
		if err := rule.action.Execute(ctx, fireContext); err != nil {
			return fmt.Errorf("execute action of rule '%s' of knowledge base '%s': %w", rule.name, k.name, err)
		}
	}

	ret.add(rule.retracts...)
	flow.deactivateUnits(rule.deactivateUnits...)
	flow.activateUnits(rule.activateUnits...)
	ret.reject(rule.inserts...)
	flow.setFocus(rule.focusUnits...)

	if rule.activationUnit != nil {
		var names []string

		for _, r := range k.activationUnits[*rule.activationUnit] {
			names = append(names, r.name)
		}

		ret.add(reject(names, rule.name)...)
	}

	return nil
}
