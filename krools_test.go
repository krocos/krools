package krools_test

import (
	"context"
	"log"
	"testing"

	"github.com/krocos/krools/v2"
)

type Content struct {
	v string
	n int
}

type Result struct {
	v string
}

type intermediateResult struct {
	v string
}

func TestSome(t *testing.T) {
	k := krools.NewKnowledgeBase("some base")

	rule := krools.NewInlineRule(
		"some rule",
		func(ctx krools.Context) (bool, error) {
			v := new(Content)
			if ctx.Get(v) && v.v == "content" && ctx.HasNot(Result{}) {
				ctx.SetLocal(intermediateResult{v: "some result to log"})
				return true, nil
			}

			return false, nil
		},
		func(ctx krools.Context) error {
			ir := new(intermediateResult)
			ctx.GetLocal(ir)

			ctx.Set(Result{v: ir.v})

			cn := new(Content)
			ctx.Get(cn)
			cn.v = "another content"
			ctx.Set(cn)

			return nil
		},
	)
	k.Add(rule)
	sr := new(SomeRule)
	k.Add(krools.NewRule("another rule", sr).Salience(1))

	s := k.NewSession()

	s.Set(&Content{v: "content"})

	err := s.FireAllRules(context.Background(), krools.RuleNameEndsWith("rule"))
	if err != nil {
		t.Fatal(err)
	}

	r := new(Result)
	if s.Get(r) {
		t.Log(r.v)
	} else {
		t.Fatal("no value")
	}

	cn := new(Content)
	s.Get(cn)
	t.Log(cn.v)
}

type SomeRule struct{}

func (s *SomeRule) When(ctx krools.Context) (bool, error) {
	c := new(Content)
	if ctx.Get(c) && c.n < 5 {
		log.Println("need to add n and n now is: ", c.n)
		return true, nil
	}

	log.Println("we do not need to add n and n is now: ", c.n)

	return false, nil
}

func (s *SomeRule) Then(ctx krools.Context) error {
	c := new(Content)
	ctx.Get(c)
	c.n++
	ctx.Set(c)
	log.Println("add n and n now is: ", c.n)
	return nil
}

type Inner struct {
	val int
}

type Outer struct {
	inner *Inner
}

func TestPtr(t *testing.T) {
	k := krools.NewKnowledgeBase("some base")
	k.Add(krools.NewInlineRule("handle pointer", func(ctx krools.Context) (bool, error) {
		o := new(Outer)
		return ctx.Get(o) && o.inner != nil && o.inner.val < 10, nil
	}, func(ctx krools.Context) error {
		if o := new(Outer); ctx.Get(o) {
			o.inner.val++
		}
		return nil
	}))

	s := k.NewSession()

	s.Set(Outer{inner: &Inner{}})

	err := s.FireAllRules(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	o := new(Outer)
	if !(s.Get(o) && o.inner != nil && o.inner.val == 10) {
		t.Fatal("unexpected counter")
	}
}
