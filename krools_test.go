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
