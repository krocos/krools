package krools

const preStartPos int = -1

type flowController struct {
	ret        *retracting
	units      map[string][]*RuleHandle
	unitsOrder []string

	pos      int
	resetPos bool
}

func newFlowController(
	ret *retracting,
	units map[string][]*RuleHandle,
	unitsOrder []string,
	deactivatedUnits []string,
) *flowController {
	c := &flowController{
		ret:        ret,
		units:      units,
		unitsOrder: unitsOrder,
		pos:        preStartPos,
	}

	c.deactivateUnits(deactivatedUnits...)

	return c
}

func (c *flowController) activateUnits(units ...string) {
	c.ret.reject(c.unitsRuleNames(units...)...)
}

func (c *flowController) deactivateUnits(units ...string) {
	c.ret.add(c.unitsRuleNames(units...)...)
}

func (c *flowController) unitsRuleNames(units ...string) []string {
	var ruleNames []string

	for _, u := range units {
		for _, r := range c.units[u] {
			ruleNames = append(ruleNames, r.name)
		}
	}

	return ruleNames
}

func (c *flowController) setFocus(units ...string) {
	c.unitsOrder = uniq(append(units, c.unitsOrder...))

	if len(units) > 0 {
		c.resetPos = true
	}
}

func (c *flowController) rules() []*RuleHandle {
	return c.units[c.unitsOrder[c.pos]]
}

func (c *flowController) more() bool {
	if c.resetPos {
		c.pos = preStartPos
		c.resetPos = false
	}

	c.pos++

	return c.pos < len(c.unitsOrder)
}
