package krools

const preStartPos int = -1

type flowController[T any] struct {
	ret        *retracting
	units      map[string][]*Rule[T]
	unitsOrder []string

	pos      int
	resetPos bool
}

func newFlowController[T any](
	ret *retracting,
	units map[string][]*Rule[T],
	unitsOrder []string,
	deactivatedUnits []string,
) *flowController[T] {
	c := &flowController[T]{
		ret:        ret,
		units:      units,
		unitsOrder: unitsOrder,
		pos:        preStartPos,
	}

	c.deactivateUnits(deactivatedUnits...)

	return c
}

func (c *flowController[T]) activateUnits(units ...string) {
	c.ret.reject(c.unitsRuleNames(units...)...)
}

func (c *flowController[T]) deactivateUnits(units ...string) {
	c.ret.add(c.unitsRuleNames(units...)...)
}

func (c *flowController[T]) unitsRuleNames(units ...string) []string {
	var ruleNames []string

	for _, u := range units {
		for _, r := range c.units[u] {
			ruleNames = append(ruleNames, r.name)
		}
	}

	return ruleNames
}

func (c *flowController[T]) setFocus(units ...string) {
	c.unitsOrder = uniq(append(units, c.unitsOrder...))
	if len(units) > 0 {
		c.resetPos = true
	}
}

func (c *flowController[T]) rules() []*Rule[T] {
	return c.units[c.unitsOrder[c.pos]]
}

func (c *flowController[T]) more() bool {
	if c.resetPos {
		c.pos = preStartPos
		c.resetPos = false
	}

	c.pos++

	return c.pos < len(c.unitsOrder)
}
