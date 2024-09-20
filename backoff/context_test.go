package backoff

import (
	"context"
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

	ctx := context.Background()
	ctx = context.WithValue(ctx, Key1{}, value)

	if ctx.Value(Key1{}) != value {
		t.Fatal("CtxStructKey can't read same value again")
	}
	if ctx.Value(Key2{}) != nil {
		t.Fatal("CtxStructKey is getting same value for different struct")
	}
}

func TestCtxStructKey_Read(t *testing.T) {
	type Key struct {
		CtxStructKey[Key, string]
	}
	ctx := context.WithValue(context.Background(), Key{}, value)
	if v := (Key{}).Must(ctx); v != value {
		t.Fatalf("CtxStructKey.Must not work properly, expected '%s' got '%v'", value, v)
	}
	if v, ok := (Key{}).Get(ctx); !ok || v != value {
		t.Fatalf("CtxStructKey.Get not work properly, got ok: '%v', value: '%v', expected: '%s'", ok, v, value)
	}
}

func TestCtxStructKey_Write(t *testing.T) {
	type Key struct {
		CtxStructKey[Key, string]
	}
	ctx := Key{}.Set(context.Background(), value)
	if ctx.Value(Key{}) != value {
		t.Fatal("CtxStructKey.Set not work properly")
	}
}
