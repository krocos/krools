package krools

import (
	"testing"
)

type S struct{}

func TestTypeContainer_Set_ByOne(t *testing.T) {
	c := newStructTypeContainer()
	i := 500

	c.Set(&S{})
	c.Set(S{})

	defer func() {
		if recover() == nil {
			t.Fatal()
		}
	}()
	c.Set(&i)
}

func TestStructType_Delete(t *testing.T) {
	c := newStructTypeContainer()
	c.Set(&S{})

	s := new(S)

	if !c.Get(s) {
		t.Fatal()
	}

	c.Delete(S{})
	if c.Get(s) {
		t.Fatal()
	}
}

type A struct{ val string }
type B struct{ val string }
type C struct{ val string }

func TestTypeContainer_Has(t *testing.T) {
	var (
		a1 = A{val: "a"}
		b1 = &B{val: "b"}
		c1 = &C{val: "c"}
	)

	var (
		a = new(A)
		b = new(B)
		c = new(C)
	)

	cn := newStructTypeContainer()

	cn.Set(a1)
	cn.Set(b1)
	cn.Set(c1)

	if cn.Get(a) && a.val != "a" {
		t.Fatal()
	}
	if cn.Get(b) && b.val != "b" {
		t.Fatal()
	}
	if cn.Get(c) && c.val != "c" {
		t.Fatal()
	}
}

func TestTypeContainer_HasNot_ViaHas(t *testing.T) {
	var (
		a1 = A{val: "a"}
		c1 = &C{val: "c"}
	)

	var (
		a = new(A)
		b = new(B)
		c = new(C)
	)

	cn := newStructTypeContainer()

	cn.Set(a1)
	cn.Set(c1)

	if cn.Get(a) && a.val != "a" {
		t.Fatal()
	}
	if !cn.Get(b) && b.val != "" {
		t.Fatal()
	}
	if cn.Get(c) && c.val != "c" {
		t.Fatal()
	}
}

func TestTypeContainer_HasNot(t *testing.T) {
	var (
		a1 = A{val: "a"}
		c1 = &C{val: "c"}
	)

	var (
		a = new(A)
		b = new(B)
		c = new(C)
	)

	cn := newStructTypeContainer()

	cn.Set(a1)
	cn.Set(c1)

	if cn.HasNot(a) {
		t.Fatal()
	}
	if !cn.HasNot(b) {
		t.Fatal()
	}
	if !cn.HasNot(B{}) {
		t.Fatal()
	}
	if cn.HasNot(c) {
		t.Fatal()
	}
}

type localType struct {
	v string
}

func TestStructType_Has_local(t *testing.T) {
	c := newStructTypeContainer()

	c.Set(localType{v: "some value"})

	l := new(localType)
	if !c.Get(l) && l.v != "some value" {
		t.Fatal()
	}
}

func TestStructTypeContainer_Handle(t *testing.T) {
	c := newStructTypeContainer()
	c.Set(localType{v: "some value"})
	if handle, ok := c.Handle(localType{}).(*localType); ok {
		t.Log(handle.v)
		handle.v = "other value"
	} else {
		t.Fatal("not found")
	}

	lv := new(localType)
	if c.Get(lv) {
		if lv.v != "other value" {
			t.Fatal("value is not changed")
		}
	}
}
