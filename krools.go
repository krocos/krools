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

type Action interface {
	Execute(ctx context.Context, fact any) error
	Describer
}

type Satisfiable interface {
	IsSatisfiedBy(ctx context.Context, candidate any) (bool, error)
	Describe() string
}

type Rule struct {
	name      string
	salience  int
	condition Satisfiable
	action    Action
	retracts  []string
}

func (r *Rule) Retracts(rules ...string) *Rule {
	r.retracts = append(r.retracts, rules...)

	return r
}

func NewRule(name string, condition Satisfiable, action Action) *Rule {
	return &Rule{name: name, condition: condition, action: action}
}

func (r *Rule) SetSalience(salience int) *Rule {
	r.salience = salience

	return r
}

func (r *Rule) Describe() string {
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

type Set struct {
	name                       string
	rules                      map[string]*Rule
	conflictResolutionStrategy int
	maxReevaluations           int
}

func NewSet(name string) *Set {
	return &Set{
		name:                       name,
		rules:                      make(map[string]*Rule),
		conflictResolutionStrategy: FireOnlyMostSalienceRule,
		maxReevaluations:           256,
	}
}

func (s *Set) Add(rule *Rule) *Set {
	s.rules[rule.name] = rule

	return s
}

func (s *Set) SetMaxReevaluations(v int) *Set {
	s.maxReevaluations = v

	return s
}

func (s *Set) Describe() string {
	rules := make([]*Rule, 0)
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

func (s *Set) FireOnlyMostSalienceRule(ctx context.Context, fact any) error {
	s.conflictResolutionStrategy = FireOnlyMostSalienceRule

	return s.fireRules(ctx, fact)
}

func (s *Set) FireAllApplicableOnce(ctx context.Context, fact any) error {
	s.conflictResolutionStrategy = FireAllApplicableOnce

	return s.fireRules(ctx, fact)
}

func (s *Set) FireMostSalienceAndReevaluate(ctx context.Context, fact any) error {
	s.conflictResolutionStrategy = FireMostSalienceAndReevaluate

	return s.fireRules(ctx, fact)
}

func (s *Set) FireAllApplicableAndReevaluate(ctx context.Context, fact any) error {
	s.conflictResolutionStrategy = FireAllApplicableAndReevaluate

	return s.fireRules(ctx, fact)
}

func (s *Set) fireRules(ctx context.Context, fact any) error {
	ret := new(retracting)

	applicable, err := s.applicableRules(ctx, fact, ret)
	if err != nil {
		return err
	}

	sortRulesConsiderSalience(applicable)

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
	retracted []string
}

func (r *retracting) Add(rules ...string) { r.retracted = append(r.retracted, rules...) }

func (r *retracting) IsRetracted(rule string) bool {
	for _, ret := range r.retracted {
		if ret == rule {
			return true
		}
	}

	return false
}

func (s *Set) applicableRules(ctx context.Context, fact any, ret *retracting) ([]*Rule, error) {
	var applicable []*Rule

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

func (s *Set) executeAction(ctx context.Context, fact any, rule *Rule, ret *retracting) error {
	if ret.IsRetracted(rule.name) {
		return nil
	}

	if err := rule.action.Execute(ctx, fact); err != nil {
		return fmt.Errorf("execute action of rule '%s' of set '%s': %w", rule.name, s.name, err)
	}

	ret.Add(rule.retracts...)

	return nil
}

func sortRulesConsiderSalience(rules []*Rule) {
	slices.SortFunc(rules, func(a, b *Rule) int {
		return cmp.Compare(a.salience, b.salience) * -1
	})
}

type ActionSet struct {
	actions []Action
}

func NewActionSet(actions ...Action) *ActionSet {
	return &ActionSet{actions: actions}
}

func (s *ActionSet) Describe() string {
	if len(s.actions) == 0 {
		return "<no actions defined>"
	}

	list := make([]string, 0)
	for _, action := range s.actions {
		list = append(list, action.Describe())
	}

	return strings.Join(list, ";\n\t\t") + ";"
}

func (s *ActionSet) Execute(ctx context.Context, fact any) error {
	for _, action := range s.actions {
		if err := action.Execute(ctx, fact); err != nil {
			return err
		}
	}

	return nil
}
