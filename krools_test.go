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

func (c *PriceGreaterThan) IsSatisfiedBy(ctx context.Context, candidate *Fact) (bool, error) {
	return candidate.Price > c.price, nil
}

type SetLowPriceAction struct{}

func (s *SetLowPriceAction) Execute(ctx context.Context, fact *Fact) error {
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
	set := krools.NewSet[*Fact]("Example set").
		Add(krools.NewRule[*Fact]("Tax for big price", priceGreater100, bigPriceAction).
			Retracts("Tax for low price").Salience(1)).
		Add(krools.NewRule[*Fact]("Tax for low price", priceGreater10, krools.NewActionStack[*Fact](
			lowPriceTaAction,
			krools.ActionFn[*Fact](func(ctx context.Context, fact *Fact) error {
				t.Log("set tax for low price")
				return nil
			}),
		)))

	err := set.FireAllRules(context.Background(), f)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if f.Tax != 10 {
		t.Error("f.Tax != 10")
	}
}

func TestAgendaGroups(t *testing.T) {
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

	set := krools.NewSet[struct{}]("some").
		SetFocus("first", "groupThatDoesNotExists", "second").
		Add(a.Salience(2)).
		Add(b.Salience(1)).
		Add(c.AgendaGroup("second")).
		Add(d.Salience(1).AgendaGroup("second")).
		Add(e.AgendaGroup("first")).
		Add(f.AgendaGroup("first"))

	err := set.FireAllRules(context.Background(), struct{}{})
	if err != nil {
		t.FailNow()
	}

	if !(order == "fedcab" || order == "efdcab") {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}

func TestActivationGroup(t *testing.T) {
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

	set := krools.NewSet[struct{}]("some").
		Add(a.ActivationGroup("g")).
		Add(b.ActivationGroup("g").Salience(1)).
		Add(c)

	err := set.FireAllRules(context.Background(), struct{}{})
	if err != nil {
		t.FailNow()
	}

	if !(order == "bc") {
		t.Fatalf("unexpected order of execution: %s", order)
	}
}
