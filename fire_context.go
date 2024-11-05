package krools

import (
	"context"
)

type fireContext struct {
	ctx context.Context
	*structTypeContainer
	rule *RuleHandle
}

func (f *fireContext) Context() context.Context {
	return f.ctx
}

func (f *fireContext) SetLocal(v any) {
	f.rule.locals.Set(v)
}

func (f *fireContext) GetLocal(v any) bool {
	return f.rule.locals.Get(v)
}

func (f *fireContext) LocalHandle(v any) any {
	return f.rule.locals.Handle(v)
}

func (f *fireContext) HasNotLocal(v any) bool {
	return f.rule.locals.HasNot(v)
}

func (f *fireContext) DeleteLocal(v any) {
	f.rule.locals.Delete(v)
}
