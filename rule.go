package krools

type Rule[T any] struct {
	name        string
	salience    int
	condition   Satisfiable[T]
	action      Action[T]
	retracts    []string
	agendaGroup string
}

const mainAgendaGroup = "MAIN"

func NewRule[T any](name string, condition Satisfiable[T], action Action[T]) *Rule[T] {
	return &Rule[T]{
		name:        name,
		condition:   condition,
		action:      action,
		retracts:    []string{name},
		agendaGroup: mainAgendaGroup,
	}
}

func (r *Rule[T]) Retracts(rules ...string) *Rule[T] {
	if len(rules) == 0 {
		rules = append(rules, r.name)
	}

	r.retracts = uniq(append(r.retracts, rules...))

	return r
}

func (r *Rule[T]) AgendaGroup(agendaGroup string) *Rule[T] {
	r.agendaGroup = agendaGroup

	return r
}

func (r *Rule[T]) DoNotAutoRetract() *Rule[T] {
	r.retracts = reject(r.retracts, r.name)

	return r
}

func (r *Rule[T]) Salience(salience int) *Rule[T] {
	r.salience = salience

	return r
}
