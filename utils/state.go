package utils

import "github.com/google/uuid"

type State struct {
	Uuid   uuid.UUID
	values map[any]any
}

func NewState() *State {
	return &State{uuid.New(), make(map[any]any)}
}

func (s *State) Put(k any, v any) {
	s.values[k] = v
}

func (s *State) Del(k any) {
	delete(s.values, k)
}

func (s *State) Get(k any) (any, bool) {
	a, ok := s.values[k]
	return a, ok
}

func GetStateValue[T any](state *State, k any) (t T, out bool) {
	if a, ok := state.Get(k); ok {
		if b, ok := a.(T); ok {
			t = b
			out = true
		}
	}
	return
}
