package krools

import (
	"fmt"
	"reflect"
)

type Container struct {
	c map[string]any
}

func NewContainer() *Container {
	return &Container{c: make(map[string]any)}
}

func (c *Container) Insert(structs ...any) {
	for i, v := range structs {
		dr := deref(reflect.ValueOf(v))

		if dr.Kind() != reflect.Struct {
			panic(fmt.Sprintf("%d argument is not a struct", i))
		}

		c.c[path(dr)] = v
	}
}

func (c *Container) Retract(structs ...any) {
	for i, v := range structs {
		dr := deref(reflect.ValueOf(v))

		if dr.Kind() != reflect.Struct {
			panic(fmt.Sprintf("%d argument is not a struct", i))
		}

		delete(c.c, path(dr))
	}
}

func (c *Container) FillIn(structs ...any) bool {
	all := true

	for i, v := range structs {
		val := reflect.ValueOf(v)
		dr := deref(val)

		if val.Kind() != reflect.Ptr {
			panic(fmt.Sprintf("%d argument is not a pointer", i))
		}

		if dr.Kind() != reflect.Struct {
			panic(fmt.Sprintf("%d argument is not a struct or is not instantiated", i))
		}

		if x, ok := c.c[path(dr)]; !ok {
			all = false
			continue
		} else {
			if dr.CanSet() {
				dr.Set(deref(reflect.ValueOf(x)))
			} else {
				panic("can't set dereferenced value")
			}
		}
	}

	return all
}

func path(dr reflect.Value) string {
	t := dr.Type()

	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

func deref(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		return deref(v)
	}

	return v
}
