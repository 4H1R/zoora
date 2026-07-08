package payment

import "fmt"

// Registry maps a gateway name to its implementation.
type Registry struct {
	gateways map[string]Gateway
}

func NewRegistry(gws ...Gateway) *Registry {
	m := make(map[string]Gateway, len(gws))
	for _, g := range gws {
		m[g.Name()] = g
	}
	return &Registry{gateways: m}
}

// ErrGatewayNotFound is returned when a name has no registered gateway.
var ErrGatewayNotFound = fmt.Errorf("payment gateway not found")

func (r *Registry) Get(name string) (Gateway, error) {
	g, ok := r.gateways[name]
	if !ok {
		return nil, ErrGatewayNotFound
	}
	return g, nil
}

func (r *Registry) Names() []string {
	out := make([]string, 0, len(r.gateways))
	for n := range r.gateways {
		out = append(out, n)
	}
	return out
}
