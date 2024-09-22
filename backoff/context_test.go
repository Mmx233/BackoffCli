package backoff

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const value = "value"

func TestCtxStructKey(t *testing.T) {
	type Key1 struct {
		CtxStructKey[Key1, string]
	}
	type Key2 struct {
		CtxStructKey[Key2, int]
	}
	ctx := context.WithValue(context.Background(), Key1{}, value)

	require.Equal(t, value, ctx.Value(Key1{}), "can't read same value again")
	require.Equal(t, nil, ctx.Value(Key2{}), "getting odd value for not exist key")
}

func TestCtxStructKey_Read(t *testing.T) {
	type Key struct {
		CtxStructKey[Key, string]
	}
	ctx := context.WithValue(context.Background(), Key{}, value)

	require.Condition(t, func() (success bool) {
		v, ok := (Key{}).Get(ctx)
		return ok && v == value
	}, "CtxStructKey.Get not work properly")

	require.Equal(t, value, (Key{}).Must(ctx), "CtxStructKey.Must not work properly")
	assert.Panics(t, func() {
		(Key{}).Must(context.Background())
	}, "should panic when key is not exist")
}

func TestCtxStructKey_Write(t *testing.T) {
	type Key struct {
		CtxStructKey[Key, string]
	}
	ctx := Key{}.Set(context.Background(), value)

	assert.Equal(t, value, (Key{}).Must(ctx), "CtxStructKey.Set not work properly")
}
