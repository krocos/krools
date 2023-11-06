package krools_test

import (
	"context"
	"fmt"
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

func (c *PriceGreaterThan) Describe() string {
	return fmt.Sprintf("price > %d", c.price)
}

type SetBigPriceAction struct{}

func (s *SetBigPriceAction) Execute(ctx context.Context, fact *Fact) error {
	fact.Tax = 10
	return nil
}

func (s *SetBigPriceAction) Describe() string {
	return "tax = 10"
}

type SetLowPriceAction struct{}

func (s *SetLowPriceAction) Execute(ctx context.Context, fact *Fact) error {
	fact.Tax = 5
	return nil
}

func (s *SetLowPriceAction) Describe() string {
	return "tax = 5"
}

func TestKrools(t *testing.T) {
	f := &Fact{Price: 102}

	priceGreater100 := NewPriceGreaterThan(100)
	priceGreater10 := NewPriceGreaterThan(10)

	bigPriceTaAction := new(SetBigPriceAction)
	lowPriceTaAction := new(SetLowPriceAction)

	set := krools.NewSet[*Fact]("Example set").
		Add(krools.NewRule[*Fact]("Tax for big price", priceGreater100, bigPriceTaAction).
			Retracts("Tax for low price").SetSalience(1)).
		Add(krools.NewRule[*Fact]("Tax for low price", priceGreater10, lowPriceTaAction))

	err := set.FireAllApplicableOnce(context.Background(), f)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// f.Tax is equal 10

	// set.Describe() will return
	// set "Example set"
	//
	//	rule "Tax for big price" salience 1
	//		retracts
	//			"Tax for low price"
	//		when
	//			price > 100
	//		then
	//			tax = 10
	//
	//	rule "Tax for low price"
	//		when
	//			price > 10
	//		then
	//			tax = 5
}
