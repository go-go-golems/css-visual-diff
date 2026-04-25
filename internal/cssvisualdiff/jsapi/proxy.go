package jsapi

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/dop251/goja"
)

const proxyIDProperty = "__cssVisualDiffProxyID"

// ProxyRegistry tracks Go backing values for Goja Proxy objects created by the
// css-visual-diff JS API. Strict lower-level APIs use it to reject raw objects
// and unwrap only handles/builders created by this package.
type ProxyRegistry struct {
	mu     sync.Mutex
	nextID int64
	items  map[int64]proxyBinding
}

type proxyBinding struct {
	Owner string
	Value any
}

func NewProxyRegistry() *ProxyRegistry {
	return &ProxyRegistry{items: map[int64]proxyBinding{}}
}

type MethodSpec struct {
	Owner string
	Hint  string
}

type ProxyMethod func(call goja.FunctionCall, receiver goja.Value) goja.Value

type ProxySpec struct {
	Owner        string
	Methods      map[string]ProxyMethod
	MethodOwners map[string]MethodSpec
}

func (s ProxySpec) availableMethods() []string {
	methods := make([]string, 0, len(s.Methods))
	for method := range s.Methods {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	return methods
}

func newProxyValue(vm *goja.Runtime, registry *ProxyRegistry, spec ProxySpec, backing any) goja.Value {
	if registry == nil {
		registry = NewProxyRegistry()
	}
	id := registry.bind(spec.Owner, backing)
	target := vm.NewObject()
	_ = target.Set(proxyIDProperty, id)

	proxy := vm.NewProxy(target, &goja.ProxyTrapConfig{
		Get: func(target *goja.Object, property string, receiver goja.Value) goja.Value {
			if property == proxyIDProperty {
				return vm.ToValue(id)
			}
			if method, ok := spec.Methods[property]; ok {
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					return method(call, receiver)
				})
			}
			if property == "toString" {
				return vm.ToValue(func() string { return fmt.Sprintf("[object %s]", spec.Owner) })
			}
			panic(unknownMethodError(vm, spec.Owner, property, spec.availableMethods(), spec.MethodOwners))
		},
	})
	return vm.ToValue(proxy)
}

func (r *ProxyRegistry) bind(owner string, value any) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	if r.items == nil {
		r.items = map[int64]proxyBinding{}
	}
	r.items[r.nextID] = proxyBinding{Owner: owner, Value: value}
	return r.nextID
}

func (r *ProxyRegistry) lookup(id int64) (proxyBinding, bool) {
	if r == nil {
		return proxyBinding{}, false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	binding, ok := r.items[id]
	return binding, ok
}

func unknownMethodError(vm *goja.Runtime, owner, property string, available []string, methodOwners map[string]MethodSpec) goja.Value {
	if methodOwner, ok := methodOwners[property]; ok && methodOwner.Owner != "" && methodOwner.Owner != owner {
		return wrongParentError(vm, owner, property, methodOwner)
	}

	message := fmt.Sprintf("%s: unknown method .%s(). Available: %s.", owner, property, strings.Join(available, ", "))
	if suggestion := closestMethod(property, available); suggestion != "" {
		message += fmt.Sprintf(" Did you mean .%s()?", suggestion)
	}
	return vm.NewTypeError(message)
}

func wrongParentError(vm *goja.Runtime, owner, property string, methodOwner MethodSpec) goja.Value {
	message := fmt.Sprintf("%s: .%s() is not available here. .%s() belongs to %s.", owner, property, property, methodOwner.Owner)
	if methodOwner.Hint != "" {
		message += " " + methodOwner.Hint
	}
	return vm.NewTypeError(message)
}

func typeMismatchError(vm *goja.Runtime, operation, expected string, got goja.Value) goja.Value {
	return vm.NewTypeError(fmt.Sprintf("%s: expected %s, got %s.", operation, expected, valueKind(got)))
}

func valueKind(value goja.Value) string {
	if value == nil || goja.IsUndefined(value) {
		return "undefined"
	}
	if goja.IsNull(value) {
		return "null"
	}
	if exported := value.Export(); exported != nil {
		return fmt.Sprintf("%T", exported)
	}
	return value.String()
}

func closestMethod(property string, available []string) string {
	best := ""
	bestDistance := 3
	for _, candidate := range available {
		distance := levenshtein(property, candidate)
		if distance < bestDistance {
			bestDistance = distance
			best = candidate
		}
	}
	return best
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return len(b)
	}
	if b == "" {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}
