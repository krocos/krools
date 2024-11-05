package krools

type KnowledgeBase struct {
	name             string
	units            map[string][]*RuleHandle
	unitsOrder       []string
	activationUnits  map[string][]*RuleHandle
	deactivatedUnits []string
}

func NewKnowledgeBase(name string) *KnowledgeBase {
	return &KnowledgeBase{
		name:            name,
		units:           make(map[string][]*RuleHandle),
		activationUnits: make(map[string][]*RuleHandle),
	}
}

func (k *KnowledgeBase) Add(rule *RuleHandle) *KnowledgeBase {
	var units []*RuleHandle

	for _, existing := range k.units[rule.unit] {
		if existing.name != rule.name {
			units = append(units, existing)
		}
	}

	k.units[rule.unit] = append(units, rule)
	k.unitsOrder = uniq(append(k.unitsOrder, rule.unit))

	if rule.activationUnit != nil {
		var activationUnits []*RuleHandle

		for _, existing := range k.activationUnits[*rule.activationUnit] {
			if existing.name != rule.name {
				activationUnits = append(activationUnits, existing)
			}
		}

		k.activationUnits[*rule.activationUnit] = append(activationUnits, rule)
	}

	return k
}

func (k *KnowledgeBase) AddUnit(unit string, rules ...*RuleHandle) *KnowledgeBase {
	for _, rule := range rules {
		k.Add(rule.Unit(unit))
	}

	return k
}

func (k *KnowledgeBase) NewSession() *Session {
	units := make(map[string][]*RuleHandle)
	for k, rr := range k.units {
		units[k] = copySliceOfRules(rr)
	}

	unitsOrder := make([]string, len(k.unitsOrder))
	copy(unitsOrder, k.unitsOrder)

	activationUnits := make(map[string][]*RuleHandle)
	for k, rr := range k.activationUnits {
		activationUnits[k] = copySliceOfRules(rr)
	}

	deactivatedUnits := make([]string, len(k.deactivatedUnits))
	copy(deactivatedUnits, k.deactivatedUnits)

	return newSession(k.name, units, unitsOrder, activationUnits, deactivatedUnits)
}
