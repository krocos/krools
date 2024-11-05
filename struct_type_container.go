package krools

import (
	"reflect"
)

type structTypeContainer struct {
	vals map[string]any
}

func newStructTypeContainer() *structTypeContainer {
	return &structTypeContainer{vals: make(map[string]any)}
}

// Set sets value in container, so passed value must be a struct or a pinter to struct.
func (c *structTypeContainer) Set(v any) {
	if v == nil {
		panic("v cannot be nil")
	}

	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		t = t.Elem()
	} else if t.Kind() == reflect.Struct {
		ptr := reflect.New(t)
		ptr.Elem().Set(reflect.ValueOf(v))
		v = ptr.Interface()
	} else {
		panic("v must be a struct or a pointer to a struct")
	}

	n := t.Name()

	if t.PkgPath() != "" {
		n = t.PkgPath() + "." + n
	}

	c.vals[n] = v
}

// Get fills passed parameter with value if such value exists and returns true, or doesn't touch value and return false.
// Passed parameter must be a pointer to a struct.
func (c *structTypeContainer) Get(v any) bool {
	if v == nil {
		panic("any v of vv cannot be nil")
	}

	t := reflect.TypeOf(v)

	if !(t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct) {
		panic("v must be a pointer to a struct")
	}

	t = t.Elem()
	n := t.Name()

	if t.PkgPath() != "" {
		n = t.PkgPath() + "." + n
	}

	if e, exists := c.vals[n]; exists {
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf(e).Elem())
	} else {
		return false
	}

	return true
}

// Handle returns a pointer to a value and if you can't convert it to desired type, so it's not found.
func (c *structTypeContainer) Handle(v any) any {
	if v == nil {
		panic("v cannot be nil")
	}

	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		t = t.Elem()
	} else if t.Kind() != reflect.Struct {
		panic("v must be a struct or a pointer to a struct")
	}

	n := t.Name()

	if t.PkgPath() != "" {
		n = t.PkgPath() + "." + n
	}

	return c.vals[n]
}

// HasNot just checks that if passed value exists in the container and does not fill passed argument, so a struct or a
// pointer to struct may be passed.
func (c *structTypeContainer) HasNot(v any) bool {
	if v == nil {
		panic("v cannot be nil")
	}

	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		t = t.Elem()
	} else if t.Kind() != reflect.Struct {
		panic("v must be a struct or a pointer to a struct")
	}

	n := t.Name()

	if t.PkgPath() != "" {
		n = t.PkgPath() + "." + n
	}

	_, exists := c.vals[n]

	return !exists
}

// Delete deletes value from container, so passed value must be a struct or a pinter to struct and concrete value
// doesn't matter.
func (c *structTypeContainer) Delete(v any) {
	if v == nil {
		panic("v cannot be nil")
	}

	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		t = t.Elem()
	} else if t.Kind() != reflect.Struct {
		panic("v must be a struct or a pointer to a struct")
	}

	n := t.Name()

	if t.PkgPath() != "" {
		n = t.PkgPath() + "." + n
	}

	if _, exists := c.vals[n]; exists {
		delete(c.vals, n)
	}
}
