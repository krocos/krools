package krools

type Rule[T any] struct {
	name      string
	salience  int
	condition Satisfiable[T]
	action    Action[T]
	retracts  []string
}

func NewRule[T any](name string, condition Satisfiable[T], action Action[T]) *Rule[T] {
	return &Rule[T]{name: name, condition: condition, action: action, retracts: []string{name}}
}

func (r *Rule[T]) Retracts(rules ...string) *Rule[T] {
	if len(rules) == 0 {
		rules = append(rules, r.name)
	}

	r.retracts = append(r.retracts, rules...)
	r.retracts = uniq(r.retracts)

	return r
}

func (r *Rule[T]) DoNotAutoRetract() *Rule[T] {
	r.retracts = reject(r.retracts, r.name)

	return r
}

func (r *Rule[T]) SetSalience(salience int) *Rule[T] {
	r.salience = salience

	return r
}

func uniq[T comparable](collection []T) []T {
	result := make([]T, 0, len(collection))
	seen := make(map[T]struct{}, len(collection))

	for _, item := range collection {
		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		result = append(result, item)
	}

	return result
}

func reject(collection []string, value string) []string {
	result := make([]string, 0)

	for _, item := range collection {
		if item != value {
			result = append(result, item)
		}
	}

	return result
}
