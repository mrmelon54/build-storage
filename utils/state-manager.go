package utils

import (
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"net/http"
	"sync"
)

type StateManager struct {
	MSync  *sync.RWMutex
	States map[uuid.UUID]*State
	Store  sessions.Store
}

func NewStateManager(sessionStore sessions.Store) *StateManager {
	return &StateManager{
		MSync:  new(sync.RWMutex),
		States: make(map[uuid.UUID]*State),
		Store:  sessionStore,
	}
}

func (m *StateManager) SessionWrapper(cb func(http.ResponseWriter, *http.Request, *State)) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		session, _ := m.Store.Get(req, "build-storage-session")
		if a, ok := session.Values["session-key"]; ok {
			if b, ok := a.(uuid.UUID); ok {
				m.MSync.RLock()
				c, ok := m.States[b]
				m.MSync.RUnlock()
				if ok {
					cb(rw, req, c)
					return
				}
			}
		}
		u := NewState()
		m.MSync.Lock()
		m.States[u.Uuid] = u
		m.MSync.Unlock()
		session.Values["session-key"] = u.Uuid
		_ = session.Save(req, rw)
		cb(rw, req, u)
	}
}
