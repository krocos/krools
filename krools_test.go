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
			Retracts("Tax for low price").SetSalience(1)).
		Add(krools.NewRule[*Fact]("Tax for low price", priceGreater10, krools.NewActionSet[*Fact](
			lowPriceTaAction,
			krools.ActionFn[*Fact](func(ctx context.Context, fact *Fact) error {
				t.Log("set tax for low price")
				return nil
			}),
		)))

	err := set.FireAllApplicableOnce(context.Background(), f)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if f.Tax != 10 {
		t.Error("f.Tax != 10")
	}
}

type fireContext struct {
	someField *string
}

func TestArrows(t *testing.T) {
	alwaysTrue := krools.ConditionFn[*fireContext](func(ctx context.Context, candidate *fireContext) (bool, error) {
		return true, nil
	})
	setFlow := func(f string) krools.ActionFn[*fireContext] {
		return func(ctx context.Context, fact *fireContext) error {
			fact.someField = &f

			return nil
		}
	}

	var gotLog string

	logAct := krools.ActionFn[*fireContext](func(ctx context.Context, fact *fireContext) error {
		if fact.someField != nil {
			gotLog += *fact.someField
		}

		return nil
	})

	first := krools.NewRule[*fireContext]("first", alwaysTrue, krools.NewActionSet[*fireContext](setFlow("a"), logAct))
	second := krools.NewRule[*fireContext]("second", alwaysTrue, krools.NewActionSet[*fireContext](setFlow("b"), logAct))
	third := krools.NewRule[*fireContext]("third", alwaysTrue, logAct)
	neverFires := krools.NewRule[*fireContext]("never fires", alwaysTrue, logAct)

	aFlow := krools.ConditionFn[*fireContext](func(ctx context.Context, fact *fireContext) (bool, error) {
		return fact.someField != nil && *fact.someField == "a", nil
	})
	bFlow := krools.ConditionFn[*fireContext](func(ctx context.Context, fact *fireContext) (bool, error) {
		return fact.someField != nil && *fact.someField == "b", nil
	})
	nFlow := krools.ConditionFn[*fireContext](func(ctx context.Context, fact *fireContext) (bool, error) {
		return fact.someField != nil && *fact.someField == "never", nil
	})

	set := krools.NewSet[*fireContext]("arrows").
		Add(first.Retracts()).
		Add(second.Retracts().SetFlowCondition(aFlow)).
		Add(third.Retracts().SetFlowCondition(bFlow)).
		Add(neverFires.SetFlowCondition(nFlow))

	fc := new(fireContext)
	if err := set.FireAllApplicableAndReevaluate(context.Background(), fc); err != nil {
		t.Fatal(err)
	}

	if gotLog != "abb" {
		t.Fatal("wrong flow value")
	}
}
