package jsapi

import (
	"fmt"
	"math"

	"github.com/dop251/goja"
)

func unwrapProxyBacking[T any](vm *goja.Runtime, registry *ProxyRegistry, operation string, value goja.Value, owner string) (*T, error) {
	binding, ok := unwrapProxyBinding(vm, registry, value)
	if !ok {
		return nil, fmt.Errorf("%s: expected %s, got %s", operation, owner, valueKind(value))
	}
	if binding.Owner != owner {
		return nil, fmt.Errorf("%s: expected %s, got %s", operation, owner, binding.Owner)
	}
	backing, ok := binding.Value.(*T)
	if !ok {
		return nil, fmt.Errorf("%s: internal proxy backing mismatch for %s", operation, owner)
	}
	return backing, nil
}

func unwrapProxyBinding(vm *goja.Runtime, registry *ProxyRegistry, value goja.Value) (proxyBinding, bool) {
	if registry == nil || value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return proxyBinding{}, false
	}
	obj := value.ToObject(vm)
	idValue := obj.Get(proxyIDProperty)
	if idValue == nil || goja.IsUndefined(idValue) || goja.IsNull(idValue) {
		return proxyBinding{}, false
	}
	idFloat := idValue.ToFloat()
	if math.Trunc(idFloat) != idFloat {
		return proxyBinding{}, false
	}
	return registry.lookup(int64(idFloat))
}

func mustUnwrapProxyBacking[T any](vm *goja.Runtime, registry *ProxyRegistry, operation string, value goja.Value, owner string) *T {
	backing, err := unwrapProxyBacking[T](vm, registry, operation, value, owner)
	if err != nil {
		panic(typeMismatchError(vm, operation, owner, value))
	}
	return backing
}
