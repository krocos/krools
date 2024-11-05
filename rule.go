package krools

type Rule interface {
	Condition
	Action
}

type Condition interface {
	When(ctx Context) (bool, error)
}

type Action interface {
	Then(ctx Context) error
}

type RuleHandle struct {
	name           string
	salience       int
	condition      Condition
	action         Action
	retracts       []string
	inserts        []string
	unit           string
	activationUnit *string
	noLoop         bool

	deactivateUnits []string
	activateUnits   []string
	focusUnits      []string

	locals *structTypeContainer
}

func NewRule(name string, rule Rule) *RuleHandle {
	return newRule(name, rule, rule)
}

func newRule(name string, condition Condition, action Action) *RuleHandle {
	return &RuleHandle{
		name:      name,
		condition: condition,
		action:    action,
		retracts:  make([]string, 0),
		unit:      UnitMAIN,
	}
}

func NewInlineRule(name string, condition ConditionFn, action ActionFn) *RuleHandle {
	return newRule(name, condition, action)
}

func copyRule(rule *RuleHandle) *RuleHandle {
	nr := &RuleHandle{
		name:            rule.name,
		salience:        rule.salience,
		condition:       rule.condition,
		action:          rule.action,
		retracts:        make([]string, len(rule.retracts)),
		inserts:         make([]string, len(rule.inserts)),
		unit:            rule.unit,
		activationUnit:  rule.activationUnit,
		noLoop:          rule.noLoop,
		deactivateUnits: make([]string, len(rule.deactivateUnits)),
		activateUnits:   make([]string, len(rule.activateUnits)),
		focusUnits:      make([]string, len(rule.focusUnits)),
	}

	copy(nr.retracts, rule.retracts)
	copy(nr.inserts, rule.inserts)

	copy(nr.deactivateUnits, rule.deactivateUnits)
	copy(nr.activateUnits, rule.activateUnits)
	copy(nr.focusUnits, rule.focusUnits)

	nr.locals = newStructTypeContainer()

	return nr
}

func copySliceOfRules(rr []*RuleHandle) []*RuleHandle {
	ns := make([]*RuleHandle, len(rr))
	for i, r := range rr {
		ns[i] = copyRule(r)
	}

	return ns
}

func (r *RuleHandle) NoLoop() *RuleHandle {
	r.noLoop = true

	return r
}

func (r *RuleHandle) Deactivate(rules ...string) *RuleHandle {
	if len(rules) == 0 {
		rules = append(rules, r.name)
	}

	r.retracts = uniq(append(r.retracts, rules...))

	return r
}

func (r *RuleHandle) Activate(rules ...string) *RuleHandle {
	if len(rules) == 0 {
		rules = append(rules, r.name)
	}

	r.inserts = uniq(append(r.inserts, rules...))

	return r
}

func (r *RuleHandle) Unit(unit string) *RuleHandle {
	r.unit = unit

	return r
}

func (r *RuleHandle) ActivationUnit(activationUnit string) *RuleHandle {
	r.activationUnit = &activationUnit

	return r
}

func (r *RuleHandle) DeactivateUnits(units ...string) *RuleHandle {
	r.deactivateUnits = uniq(append(r.deactivateUnits, units...))

	return r
}

func (r *RuleHandle) ActivateUnits(units ...string) *RuleHandle {
	r.activateUnits = uniq(append(r.activateUnits, units...))

	return r
}

func (r *RuleHandle) SetFocus(units ...string) *RuleHandle {
	r.focusUnits = uniq(append(r.focusUnits, units...))

	return r
}

func (r *RuleHandle) Salience(salience int) *RuleHandle {
	r.salience = salience

	return r
}
