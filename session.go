package krools

import (
	"context"
	"errors"
	"fmt"
)

type Session struct {
	*structTypeContainer

	knowledgeBaseName string
	units             map[string][]*RuleHandle
	unitsOrder        []string
	activationUnits   map[string][]*RuleHandle
	deactivatedUnits  []string
	maxReevaluations  int
}

func newSession(
	knowledgeBaseName string,
	units map[string][]*RuleHandle,
	unitsOrder []string,
	activationUnits map[string][]*RuleHandle,
	deactivatedUnits []string,
) *Session {
	return &Session{
		structTypeContainer: newStructTypeContainer(),

		knowledgeBaseName: knowledgeBaseName,
		units:             units,
		unitsOrder:        unitsOrder,
		activationUnits:   activationUnits,
		deactivatedUnits:  deactivatedUnits,
		maxReevaluations:  65535,
	}
}

func (s *Session) SetMaxReevaluations(v int) *Session {
	s.maxReevaluations = v

	return s
}

func (s *Session) SetFocus(units ...string) *Session {
	s.unitsOrder = uniq(append(units, s.unitsOrder...))

	return s
}

func (s *Session) SetDeactivatedUnits(units ...string) *Session {
	s.deactivatedUnits = uniq(append(units, s.deactivatedUnits...))

	return s
}

func (s *Session) Clear() {
	s.structTypeContainer = newStructTypeContainer()
}

func (s *Session) FireAllRules(ctx context.Context, options ...any) error {
	var ruleFilters []Filter

	for _, option := range options {
		if v, ok := option.(Filter); ok {
			ruleFilters = append(ruleFilters, v)
		}
	}

	fc := &fireContext{
		ctx:                 ctx,
		structTypeContainer: s.structTypeContainer,
	}

	return s.fire(fc, ruleFilters...)
}

func (s *Session) fire(ctx *fireContext, ruleFilters ...Filter) error {
	ret := newRetracting()
	flow := newFlowController(ret, s.units, s.unitsOrder, s.deactivatedUnits)

	var reevaluations int

	for flow.more() {
		applicable, err := s.applicableRules(ctx, flow.rules(), ret, false, ruleFilters...)
		if err != nil {
			return err
		}

		for len(applicable) > 0 {
			for _, rule := range applicable {
				if err = s.executeAction(ctx, rule, ret, flow); err != nil {
					return err
				}
			}

			applicable, err = s.applicableRules(ctx, flow.rules(), ret, true, ruleFilters...)
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

func (s *Session) applicableRules(
	ctx *fireContext,
	rules []*RuleHandle,
	ret *retracting,
	discardNoLoop bool,
	filters ...Filter,
) ([]*RuleHandle, error) {
	var applicable []*RuleHandle

loop:
	for _, rule := range rules {
		if ret.isRetracted(rule.name) {
			continue
		}

		if discardNoLoop && rule.noLoop {
			continue
		}

		for i, filter := range filters {
			ok, err := filter.IsSatisfiedBy(ctx.Context(), rule)
			if err != nil {
				return nil, fmt.Errorf("verify that rule '%s' of knowledge base '%s' is satisfies filter %d: %w", rule.name, s.knowledgeBaseName, i, err)
			}

			if !ok {
				continue loop
			}
		}

		satisfied := true

		if rule.condition != nil {
			var err error
			err = func() error {
				ctx.rule = rule
				defer func() { ctx.rule = nil }()

				satisfied, err = rule.condition.When(ctx)
				if err != nil {
					return fmt.Errorf("verify that condition of rule '%s' of knowledge base '%s' is satisfied by fire context: %w", rule.name, s.knowledgeBaseName, err)
				}

				return nil
			}()

			if err != nil {
				return nil, err
			}
		}

		if satisfied {
			applicable = append(applicable, rule)
		}
	}

	sortRulesConsiderSalience(applicable)

	return applicable, nil
}

func (s *Session) executeAction(ctx *fireContext, rule *RuleHandle, ret *retracting, flow *flowController) error {
	if ret.isRetracted(rule.name) {
		return nil
	}

	if rule.action != nil {
		if err := func() error {
			ctx.rule = rule
			defer func() {
				ctx.rule.locals = newStructTypeContainer()
				ctx.rule = nil
			}()

			if err := rule.action.Then(ctx); err != nil {
				return fmt.Errorf("execute action of rule '%s' of knowledge base '%s': %w", rule.name, s.knowledgeBaseName, err)
			}

			return nil
		}(); err != nil {
			return err
		}
	}

	ret.add(rule.retracts...)
	flow.deactivateUnits(rule.deactivateUnits...)
	flow.activateUnits(rule.activateUnits...)
	ret.reject(rule.inserts...)
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
