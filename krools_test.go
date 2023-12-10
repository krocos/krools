package krools_test

import (
	"context"
	"testing"

	"github.com/krocos/krools"
)

type Fact struct {
	Price int
	Tax   int
}

type PriceGreaterThan struct {
	price int
}

func NewPriceGreaterThan(price int) *PriceGreaterThan {
	return &PriceGreaterThan{price: price}
}

func (c *PriceGreaterThan) IsSatisfiedBy(_ context.Context, candidate *Fact) (bool, error) {
	return candidate.Price > c.price, nil
}

type SetLowPriceAction struct{}

func (s *SetLowPriceAction) Execute(_ context.Context, fact *Fact) error {
	fact.Tax = 5
	return nil
}

func TestKrools(t *testing.T) {
	f := &Fact{Price: 102}

	priceGreater100 := krools.ConditionFn[*Fact](func(ctx context.Context, candidate *Fact) (bool, error) {
		return candidate.Price > 100, nil
	})
	priceGreater10 := NewPriceGreaterThan(10)

	lowPriceTaAction := new(SetLowPriceAction)

	bigPriceAction := krools.ActionFn[*Fact](func(ctx context.Context, fact *Fact) error {
		fact.Tax = 10
		return nil
	})
	k := krools.NewKnowledgeBase[*Fact]("Example set").
		Add(krools.NewRule[*Fact]("Tax for big price", priceGreater100, bigPriceAction).
			Retracts("Tax for low price").Salience(1)).
		Add(krools.NewRule[*Fact]("Tax for low price", priceGreater10, krools.NewActionStack[*Fact](
			lowPriceTaAction,
			krools.ActionFn[*Fact](func(ctx context.Context, fact *Fact) error {
				t.Log("set tax for low price")
				return nil
			}),
		)))

	err := k.FireAllRules(context.Background(), f)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if f.Tax != 10 {
		t.Error("f.Tax != 10")
	}
}

func TestUnit(t *testing.T) {
	alwaysTrue := krools.ConditionFn[struct{}](func(ctx context.Context, candidate struct{}) (bool, error) {
		return true, nil
	})

	var order string

	appendAction := func(v string) krools.Action[struct{}] {
		return krools.ActionFn[struct{}](func(ctx context.Context, fact struct{}) error {
			order += v
			return nil
		})
	}

	a := krools.NewRule[struct{}]("a", alwaysTrue, appendAction("a"))
	b := krools.NewRule[struct{}]("b", alwaysTrue, appendAction("b"))
	c := krools.NewRule[struct{}]("c", alwaysTrue, appendAction("c"))
	d := krools.NewRule[struct{}]("d", alwaysTrue, appendAction("d"))
	e := krools.NewRule[struct{}]("e", alwaysTrue, appendAction("e"))
	f := krools.NewRule[struct{}]("f", alwaysTrue, appendAction("f"))

	k := krools.NewKnowledgeBase[struct{}]("some").
		SetFocus(
			"first",
			"groupThatDoesNotExists",
			"second",
		).
		Add(a.Salience(2)).
		Add(b.Salience(1)).
		Add(c.Unit("second")).
		Add(d.Salience(1).Unit("second")).
		Add(e.Unit("first")).
		Add(f.Unit("first"))

	err := k.FireAllRules(context.Background(), struct{}{})
	if err != nil {
		t.FailNow()
	}

	if order != "efdcab" {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}

func TestActivationUnit(t *testing.T) {
	alwaysTrue := krools.ConditionFn[struct{}](func(ctx context.Context, candidate struct{}) (bool, error) {
		return true, nil
	})

	var order string

	appendAction := func(v string) krools.Action[struct{}] {
		return krools.ActionFn[struct{}](func(ctx context.Context, fact struct{}) error {
			order += v
			return nil
		})
	}

	a := krools.NewRule[struct{}]("a", alwaysTrue, appendAction("a"))
	b := krools.NewRule[struct{}]("b", alwaysTrue, appendAction("b"))
	c := krools.NewRule[struct{}]("c", alwaysTrue, appendAction("c"))

	k := krools.NewKnowledgeBase[struct{}]("some").
		Add(a.ActivationUnit("g")).
		Add(b.ActivationUnit("g").Salience(1)).
		Add(c)

	err := k.FireAllRules(context.Background(), struct{}{})
	if err != nil {
		t.FailNow()
	}

	if !(order == "bc") {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}

func TestEmptyCondition(t *testing.T) {
	var order string

	appendAction := func(v string) krools.Action[struct{}] {
		return krools.ActionFn[struct{}](func(ctx context.Context, fact struct{}) error {
			order += v
			return nil
		})
	}

	a := krools.NewRule[struct{}]("a", nil, appendAction("a"))
	b := krools.NewRule[struct{}]("b", nil, appendAction("b"))
	c := krools.NewRule[struct{}]("c", nil, appendAction("c"))

	k := krools.NewKnowledgeBase[struct{}]("some").
		Add(a).
		Add(b).
		Add(c)

	err := k.FireAllRules(context.Background(), struct{}{})
	if err != nil {
		t.FailNow()
	}

	if order != "abc" {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}

func TestRuleFilters(t *testing.T) {
	var order string

	appendAction := func(v string) krools.Action[struct{}] {
		return krools.ActionFn[struct{}](func(ctx context.Context, fact struct{}) error {
			order += v
			return nil
		})
	}

	a := krools.NewRule[struct{}]("a - test", nil, appendAction("a"))
	b := krools.NewRule[struct{}]("b", nil, appendAction("b"))
	c := krools.NewRule[struct{}]("c", nil, appendAction("c"))
	d := krools.NewRule[struct{}]("d", nil, appendAction("d"))

	k := krools.NewKnowledgeBase[struct{}]("some").
		Add(a).
		Add(b).
		Add(c).
		Add(d)

	err := k.FireAllRules(
		context.Background(),
		struct{}{},
		krools.RuleNameMustNotContainsAny[struct{}]("b"),
		krools.RuleNameMustNotContainsAny[struct{}]("c"),
		krools.RuleNameMustContainsAny[struct{}]("test"),
	)
	if err != nil {
		t.FailNow()
	}

	if order != "a" {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}

func TestFlow(t *testing.T) {
	var order string

	appendAction := func(v string) krools.Action[struct{}] {
		return krools.ActionFn[struct{}](func(ctx context.Context, fact struct{}) error {
			order += v
			return nil
		})
	}

	a := krools.NewRule[struct{}]("a", nil, appendAction("a"))
	b := krools.NewRule[struct{}]("b", nil, appendAction("b"))
	c := krools.NewRule[struct{}]("c", nil, appendAction("c"))
	d := krools.NewRule[struct{}]("d", nil, appendAction("d"))
	e := krools.NewRule[struct{}]("e", nil, appendAction("e"))
	f := krools.NewRule[struct{}]("f", nil, appendAction("f"))
	g := krools.NewRule[struct{}]("g", nil, appendAction("g"))

	k := krools.NewKnowledgeBase[struct{}]("some").
		SetFocus(
			"first",
			"second",
			"third",
		).
		SetDeactivatedUnits(
			"second",
			"first",
			"optional",
		).
		Add(a.Unit("third")).
		Add(b.Unit("third").ActivateUnits("optional")).
		Add(c.Unit("second")).
		Add(d.Unit("second").ActivateUnits("first").SetFocus("first")).
		Add(e.Unit("first").Salience(-1)).
		Add(f.Unit("first")).
		Add(g.Unit("optional").ActivateUnits("second").SetFocus("second"))

	err := k.FireAllRules(context.Background(), struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	if order != "abgcdfe" {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}

func TestFlowRuleDeactivateUnits(t *testing.T) {
	var order string

	appendAction := func(v string) krools.Action[struct{}] {
		return krools.ActionFn[struct{}](func(ctx context.Context, fact struct{}) error {
			order += v
			return nil
		})
	}

	falseCond := krools.ConditionFn[struct{}](func(ctx context.Context, candidate struct{}) (bool, error) {
		return false, nil
	})

	a := krools.NewRule[struct{}]("a", nil, appendAction("a"))
	b := krools.NewRule[struct{}]("b", falseCond, appendAction("b"))
	c := krools.NewRule[struct{}]("c", nil, appendAction("c"))
	d := krools.NewRule[struct{}]("d", nil, appendAction("d"))
	e := krools.NewRule[struct{}]("e", nil, appendAction("e"))
	f := krools.NewRule[struct{}]("f", nil, appendAction("f"))
	g := krools.NewRule[struct{}]("g", nil, appendAction("g"))
	h := krools.NewRule[struct{}]("h", nil, appendAction("h"))
	i := krools.NewRule[struct{}]("i", nil, appendAction("i"))

	k := krools.NewKnowledgeBase[struct{}]("some").
		SetDeactivatedUnits("one", "two").
		Add(a.DeactivateUnits("last")).
		Add(b.ActivateUnits("one")).
		Add(c.Unit("one")).
		Add(d.Unit("one")).
		Add(e).
		Add(f.ActivateUnits("two")).
		Add(g.Unit("two")).
		Add(h.Unit("last")).
		Add(i.Unit("last"))

	err := k.FireAllRules(context.Background(), struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	if order != "aefg" {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}

func TestRuleFilter_RunOnlyRulesFromUnits(t *testing.T) {
	var order string

	appendAction := func(v string) krools.Action[struct{}] {
		return krools.ActionFn[struct{}](func(ctx context.Context, fact struct{}) error {
			order += v
			return nil
		})
	}

	a := krools.NewRule[struct{}]("a", nil, appendAction("a"))
	b := krools.NewRule[struct{}]("b", nil, appendAction("b"))
	c := krools.NewRule[struct{}]("c", nil, appendAction("c"))
	d := krools.NewRule[struct{}]("d", nil, appendAction("d"))

	k := krools.NewKnowledgeBase[struct{}]("some").
		Add(a.Unit("C")).
		Add(b).
		Add(c.Unit("A")).
		Add(d.Unit("B"))

	err := k.FireAllRules(context.Background(), struct{}{},
		krools.RunOnlyUnits[struct{}](krools.UnitMAIN, "A"))
	if err != nil {
		t.Fatal(err)
	}

	if order != "bc" {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}

func TestRuleFilter_RuleNameStartsWith(t *testing.T) {
	var order string

	appendAction := func(v string) krools.Action[struct{}] {
		return krools.ActionFn[struct{}](func(ctx context.Context, fact struct{}) error {
			order += v
			return nil
		})
	}

	a := krools.NewRule[struct{}]("aabbcc", nil, appendAction("a"))
	b := krools.NewRule[struct{}]("bbccdd", nil, appendAction("b"))
	c := krools.NewRule[struct{}]("ccddee", nil, appendAction("c"))
	d := krools.NewRule[struct{}]("ddeeff", nil, appendAction("d"))

	k := krools.NewKnowledgeBase[struct{}]("some").
		Add(a.Unit("C")).
		Add(b).
		Add(c.Unit("A")).
		Add(d.Unit("B"))

	err := k.FireAllRules(context.Background(), struct{}{}, krools.RuleNameStartsWith[struct{}]("bb"))
	if err != nil {
		t.Fatal(err)
	}

	if order != "b" {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}

func TestRuleFilter_RuleNameEndsWith(t *testing.T) {
	var order string

	appendAction := func(v string) krools.Action[struct{}] {
		return krools.ActionFn[struct{}](func(ctx context.Context, fact struct{}) error {
			order += v
			return nil
		})
	}

	a := krools.NewRule[struct{}]("aabbee", nil, appendAction("a"))
	b := krools.NewRule[struct{}]("bbccdd", nil, appendAction("b"))
	c := krools.NewRule[struct{}]("ccddee", nil, appendAction("c"))
	d := krools.NewRule[struct{}]("ddeeff", nil, appendAction("d"))

	k := krools.NewKnowledgeBase[struct{}]("some").
		Add(a.Unit("C")).
		Add(b).
		Add(c.Unit("A")).
		Add(d.Unit("B"))

	err := k.FireAllRules(context.Background(), struct{}{}, krools.RuleNameEndsWith[struct{}]("ee"))
	if err != nil {
		t.Fatal(err)
	}

	if order != "ac" {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}

func TestRule_Inserts(t *testing.T) {
	type fc struct {
		counter int
	}

	var order string

	appendAction := func(v string) krools.Action[*fc] {
		return krools.ActionFn[*fc](func(ctx context.Context, fact *fc) error {
			order += v
			return nil
		})
	}

	counterLowerThanOne := krools.ConditionFn[*fc](func(ctx context.Context, candidate *fc) (bool, error) {
		return candidate.counter < 1, nil
	})
	counterPlusOne := krools.ActionFn[*fc](func(ctx context.Context, fireContext *fc) error {
		fireContext.counter++
		return nil
	})

	a := krools.NewRule[*fc]("a", nil, appendAction("a"))
	b := krools.NewRule[*fc]("b", nil, appendAction("b"))
	c := krools.NewRule[*fc]("c", counterLowerThanOne, krools.NewActionStack[*fc](
		counterPlusOne,
		appendAction("c"),
	))
	d := krools.NewRule[*fc]("d", nil, appendAction("d"))
	e := krools.NewRule[*fc]("e", nil, appendAction("e"))

	k := krools.NewKnowledgeBase[*fc]("some").
		Add(a).
		Add(b).
		Add(c.Inserts("b")).
		Add(d).
		Add(e.Unit("next"))

	fireContext := new(fc)

	err := k.FireAllRules(context.Background(), fireContext)
	if err != nil {
		t.Fatal(err)
	}

	if order != "abcdbe" {
		t.Fatalf("unexpected order of execution: %s", order)
	}

	if fireContext.counter != 1 {
		t.Fatalf("unexpected counter value: %d", fireContext.counter)
	}
}

func TestKnowledgeBase_AddUnit(t *testing.T) {
	var order string

	appendAction := func(v string) krools.Action[struct{}] {
		return krools.ActionFn[struct{}](func(ctx context.Context, fact struct{}) error {
			order += v
			return nil
		})
	}

	a := krools.NewRule[struct{}]("a", nil, appendAction("a"))
	b := krools.NewRule[struct{}]("b", nil, appendAction("b"))
	c := krools.NewRule[struct{}]("c", nil, appendAction("c"))
	d := krools.NewRule[struct{}]("d", nil, appendAction("d"))
	e := krools.NewRule[struct{}]("e", nil, appendAction("e"))
	f := krools.NewRule[struct{}]("f", nil, appendAction("f"))
	g := krools.NewRule[struct{}]("g", nil, appendAction("g"))

	k := krools.NewKnowledgeBase[struct{}]("add unit").
		AddUnit("first", d, e, f, g).
		AddUnit("second", a, b, c)

	if err := k.FireAllRules(context.Background(), struct{}{}); err != nil {
		t.Fatal(err)
	}

	if order != "defgabc" {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}

func TestKnowledgeBase_AddUnit_ActivateDeactivatedUnit(t *testing.T) {
	var order string

	appendAction := func(v string) krools.Action[struct{}] {
		return krools.ActionFn[struct{}](func(ctx context.Context, fact struct{}) error {
			order += v
			return nil
		})
	}

	a := krools.NewRule[struct{}]("a", nil, appendAction("a"))
	b := krools.NewRule[struct{}]("b", nil, appendAction("b"))
	c := krools.NewRule[struct{}]("c", nil, appendAction("c"))
	d := krools.NewRule[struct{}]("d", nil, appendAction("d"))
	e := krools.NewRule[struct{}]("e", nil, appendAction("e"))
	f := krools.NewRule[struct{}]("f", nil, appendAction("f"))
	g := krools.NewRule[struct{}]("g", nil, appendAction("g"))

	k := krools.NewKnowledgeBase[struct{}]("add unit").
		AddUnit("first", d, e, f, g.DeactivateUnits("second")).
		AddUnit("second", a, b, c)

	if err := k.FireAllRules(context.Background(), struct{}{}); err != nil {
		t.Fatal(err)
	}

	if order != "defg" {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}

func TestKnowledgeBase_AddUnit_Deactivation(t *testing.T) {
	var order string

	appendAction := func(v string) krools.Action[struct{}] {
		return krools.ActionFn[struct{}](func(ctx context.Context, fact struct{}) error {
			order += v
			return nil
		})
	}

	a := krools.NewRule[struct{}]("a", nil, appendAction("a"))
	b := krools.NewRule[struct{}]("b", nil, appendAction("b"))
	c := krools.NewRule[struct{}]("c", nil, appendAction("c"))
	d := krools.NewRule[struct{}]("d", nil, appendAction("d"))
	e := krools.NewRule[struct{}]("e", nil, appendAction("e"))
	f := krools.NewRule[struct{}]("f", nil, appendAction("f"))
	g := krools.NewRule[struct{}]("g", nil, appendAction("g"))

	k := krools.NewKnowledgeBase[struct{}]("add unit").
		SetDeactivatedUnits("second").
		AddUnit("first", d, e, f, g.ActivationUnit("second")).
		AddUnit("second", a, b, c)

	if err := k.FireAllRules(context.Background(), struct{}{}); err != nil {
		t.Fatal(err)
	}

	if order != "defg" {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}
