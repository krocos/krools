package krools

type retracting struct {
	retracted map[string]struct{}
}

func newRetracting() *retracting {
	return &retracting{retracted: make(map[string]struct{})}
}

func (r *retracting) add(rules ...string) {
	for _, rule := range rules {
		r.retracted[rule] = struct{}{}
	}
}

func (r *retracting) isRetracted(rule string) bool {
	_, ok := r.retracted[rule]

	return ok
}
