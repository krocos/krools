package krools

type Rule[T any] struct {
	name           string
	salience       int
	condition      Condition[T]
	action         Action[T]
	retracts       []string
	inserts        []string
	unit           string
	activationUnit *string
	noLoop         bool

	deactivateUnits []string
	activateUnits   []string
	focusUnits      []string
}

func NewRule[T any](name string, condition Condition[T], action Action[T]) *Rule[T] {
	return &Rule[T]{
		name:      name,
		condition: condition,
		action:    action,
		retracts:  []string{name},
		unit:      UnitMAIN,
	}
}

func (r *Rule[T]) NoLoop() *Rule[T] {
	r.noLoop = true

	return r
}

func (r *Rule[T]) Retract(rules ...string) *Rule[T] {
	if len(rules) == 0 {
		rules = append(rules, r.name)
	}

	r.retracts = uniq(append(r.retracts, rules...))

	return r
}

func (r *Rule[T]) Insert(rules ...string) *Rule[T] {
	if len(rules) == 0 {
		rules = append(rules, r.name)
	}

	r.inserts = uniq(append(r.inserts, rules...))

	return r
}

func (r *Rule[T]) Unit(unit string) *Rule[T] {
	r.unit = unit

	return r
}

func (r *Rule[T]) ActivationUnit(activationUnit string) *Rule[T] {
	r.activationUnit = &activationUnit

	return r
}

func (r *Rule[T]) DeactivateUnits(units ...string) *Rule[T] {
	r.deactivateUnits = uniq(append(r.deactivateUnits, units...))

	return r
}

func (r *Rule[T]) ActivateUnits(units ...string) *Rule[T] {
	r.activateUnits = uniq(append(r.activateUnits, units...))

	return r
}

func (r *Rule[T]) SetFocus(units ...string) *Rule[T] {
	r.focusUnits = uniq(append(r.focusUnits, units...))

	return r
}

func (r *Rule[T]) Salience(salience int) *Rule[T] {
	r.salience = salience

	return r
}
