package krools

import (
	"bufio"
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	FireOnlyMostSalienceRule = iota
	FireAllApplicableOnce
	FireMostSalienceAndReevaluate
	FireAllApplicableAndReevaluate
)

type Describer interface {
	Describe() string
}

type Action[T any] interface {
	Execute(ctx context.Context, fact T) error
	Describer
}

type Satisfiable[T any] interface {
	IsSatisfiedBy(ctx context.Context, candidate T) (bool, error)
	Describe() string
}

type Rule[T any] struct {
	name      string
	salience  int
	condition Satisfiable[T]
	action    Action[T]
	retracts  []string
}

// Retracts add passed rule names to retract (to not fire in next and current
// evaluation) if action of rule fired. If no names passed it add itself name.
func (r *Rule[T]) Retracts(rules ...string) *Rule[T] {
	if len(rules) == 0 {
		rules = append(rules, r.name)
	}

	r.retracts = append(r.retracts, rules...)

	return r
}

func NewRule[T any](name string, condition Satisfiable[T], action Action[T]) *Rule[T] {
	return &Rule[T]{name: name, condition: condition, action: action}
}

func (r *Rule[T]) SetSalience(salience int) *Rule[T] {
	r.salience = salience

	return r
}

func (r *Rule[T]) Describe() string {
	var b strings.Builder
	if r.salience != 0 {
		b.WriteString(fmt.Sprintf("rule \"%s\" salience %d\n", r.name, r.salience))
	} else {
		b.WriteString(fmt.Sprintf("rule \"%s\"\n", r.name))
	}

	if len(r.retracts) > 0 {
		b.WriteString("\tretracts\n")
		b.WriteString(fmt.Sprintf("\t\t%s\n", fmt.Sprintf("\"%s\"", strings.Join(r.retracts, "\",\n\t\t\""))))
	}

	b.WriteString("\twhen\n")
	b.WriteString(fmt.Sprintf("\t\t%s\n", r.condition.Describe()))
	b.WriteString("\tthen\n")
	b.WriteString(fmt.Sprintf("\t\t%s", r.action.Describe()))

	return b.String()
}

type Set[T any] struct {
	name                       string
	rules                      map[string]*Rule[T]
	conflictResolutionStrategy int
	maxReevaluations           int
}

func NewSet[T any](name string) *Set[T] {
	return &Set[T]{
		name:                       name,
		rules:                      make(map[string]*Rule[T]),
		conflictResolutionStrategy: FireOnlyMostSalienceRule,
		maxReevaluations:           256,
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

func (s *Set[T]) Describe() string {
	rules := make([]*Rule[T], 0)
	for _, rule := range s.rules {
		rules = append(rules, rule)
	}

	sortRulesConsiderSalience(rules)

	list := make([]string, 0)
	for _, rule := range rules {
		list = append(list, rule.Describe())
	}

	var prefixed strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(strings.Join(list, "\n\n")))
	for scanner.Scan() {
		prefixed.WriteString(fmt.Sprintf("\t%s\n", scanner.Text()))
	}

	return fmt.Sprintf("set \"%s\"\n\n%s", s.name, prefixed.String())
}

func (s *Set[T]) FireOnlyMostSalienceRule(ctx context.Context, fact T) error {
	s.conflictResolutionStrategy = FireOnlyMostSalienceRule

	return s.fireRules(ctx, fact)
}

func (s *Set[T]) FireAllApplicableOnce(ctx context.Context, fact T) error {
	s.conflictResolutionStrategy = FireAllApplicableOnce

	return s.fireRules(ctx, fact)
}

func (s *Set[T]) FireMostSalienceAndReevaluate(ctx context.Context, fact T) error {
	s.conflictResolutionStrategy = FireMostSalienceAndReevaluate

	return s.fireRules(ctx, fact)
}

func (s *Set[T]) FireAllApplicableAndReevaluate(ctx context.Context, fact T) error {
	s.conflictResolutionStrategy = FireAllApplicableAndReevaluate

	return s.fireRules(ctx, fact)
}

func (s *Set[T]) fireRules(ctx context.Context, fact T) error {
	ret := newRetracting()

	applicable, err := s.applicableRules(ctx, fact, ret)
	if err != nil {
		return err
	}

	sortRulesConsiderSalience[T](applicable)

	switch s.conflictResolutionStrategy {
	case FireOnlyMostSalienceRule:
		if len(applicable) > 0 {
			if err = s.executeAction(ctx, fact, applicable[0], ret); err != nil {
				return err
			}
		}
	case FireAllApplicableOnce:
		for _, rule := range applicable {
			if err = s.executeAction(ctx, fact, rule, ret); err != nil {
				return err
			}
		}
	case FireMostSalienceAndReevaluate, FireAllApplicableAndReevaluate:
		var reevaluations int

		for len(applicable) > 0 {
			switch s.conflictResolutionStrategy {
			case FireMostSalienceAndReevaluate:
				if err = s.executeAction(ctx, fact, applicable[0], ret); err != nil {
					return err
				}
			case FireAllApplicableAndReevaluate:
				for _, rule := range applicable {
					if err = s.executeAction(ctx, fact, rule, ret); err != nil {
						return err
					}
				}
			}

			applicable, err = s.applicableRules(ctx, fact, ret)
			if err != nil {
				return err
			}

			sortRulesConsiderSalience(applicable)

			reevaluations++

			if reevaluations > s.maxReevaluations {
				return errors.New("too much reevaluations")
			}
		}
	}

	return nil
}

type retracting struct {
	retracted map[string]struct{}
}

func newRetracting() *retracting {
	return &retracting{retracted: make(map[string]struct{})}
}

func (r *retracting) Add(rules ...string) {
	for _, rule := range rules {
		r.retracted[rule] = struct{}{}
	}
}

func (r *retracting) IsRetracted(rule string) bool {
	_, ok := r.retracted[rule]

	return ok
}

func (s *Set[T]) applicableRules(ctx context.Context, fact T, ret *retracting) ([]*Rule[T], error) {
	var applicable []*Rule[T]

	for _, rule := range s.rules {
		if ret.IsRetracted(rule.name) {
			continue
		}

		satisfied, err := rule.condition.IsSatisfiedBy(ctx, fact)
		if err != nil {
			return nil, fmt.Errorf("verify is condition of rule '%s' of set '%s' is satisfied by fact: %w", rule.name, s.name, err)
		}

		if satisfied {
			applicable = append(applicable, rule)
		}
	}

	return applicable, nil
}

func (s *Set[T]) executeAction(ctx context.Context, fact T, rule *Rule[T], ret *retracting) error {
	if ret.IsRetracted(rule.name) {
		return nil
	}

	if err := rule.action.Execute(ctx, fact); err != nil {
		return fmt.Errorf("execute action of rule '%s' of set '%s': %w", rule.name, s.name, err)
	}

	ret.Add(rule.retracts...)

	return nil
}

func sortRulesConsiderSalience[T any](rules []*Rule[T]) {
	slices.SortFunc(rules, func(a, b *Rule[T]) int {
		return cmp.Compare(a.salience, b.salience) * -1
	})
}

type ActionSet[T any] struct {
	actions []Action[T]
}

func NewActionSet[T any](actions ...Action[T]) *ActionSet[T] {
	return &ActionSet[T]{actions: actions}
}

func (s *ActionSet[T]) Describe() string {
	if len(s.actions) == 0 {
		return "<no actions defined>"
	}

	list := make([]string, 0)
	for _, action := range s.actions {
		list = append(list, action.Describe())
	}

	return strings.Join(list, ";\n\t\t") + ";"
}

func (s *ActionSet[T]) Execute(ctx context.Context, fact T) error {
	for _, action := range s.actions {
		if err := action.Execute(ctx, fact); err != nil {
			return err
		}
	}

	return nil
}
