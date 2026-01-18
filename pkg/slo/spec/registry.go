package spec

import "fmt"

type Registry struct {
	items map[string]SLISpec
}

func NewRegistry() *Registry {
	return &Registry{items: map[string]SLISpec{}}
}

func (r *Registry) Register(s SLISpec) error {
	if s.ID == "" {
		return fmt.Errorf("sli spec id is required")
	}
	if _, exists := r.items[s.ID]; exists {
		return fmt.Errorf("sli spec already registered: %s", s.ID)
	}
	r.items[s.ID] = s
	return nil
}

func (r *Registry) MustRegister(s SLISpec) {
	if err := r.Register(s); err != nil {
		panic(err)
	}
}

func (r *Registry) Get(id string) (SLISpec, bool) {
	s, ok := r.items[id]
	return s, ok
}

func (r *Registry) List() []SLISpec {
	out := make([]SLISpec, 0, len(r.items))
	for _, s := range r.items {
		out = append(out, s)
	}
	return out
}
