package backoff

import (
	"context"
	"fmt"
)

type CtxStructKey[Key, Value any] struct{}

func (CtxStructKey[Key, Value]) Set(ctx context.Context, value Value) context.Context {
	var key Key
	return context.WithValue(ctx, key, value)
}
func (CtxStructKey[Key, Value]) Get(ctx context.Context) (Value, bool) {
	var (
		key   Key
		value Value
		ok    bool
	)
	anyValue := ctx.Value(key)
	if anyValue != nil {
		value, ok = anyValue.(Value)
	}
	return value, ok
}
func (c CtxStructKey[Key, Value]) Must(ctx context.Context) Value {
	value, ok := c.Get(ctx)
	if !ok {
		var key Key
		panic(fmt.Sprintf("%T is required in context", key))
	}
	return value
}

type CtxResetWait struct {
	CtxStructKey[CtxResetWait, chan struct{}]
}
type CtxCancelFn struct {
	CtxStructKey[CtxCancelFn, context.CancelFunc]
}
