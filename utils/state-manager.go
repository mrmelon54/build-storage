package utils

import (
	"encoding/gob"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"sync"
)

type StateManager struct {
	MSync  *sync.RWMutex
	States map[uuid.UUID]*State
	Store  sessions.Store
}

func NewStateManager(sessionStore sessions.Store) *StateManager {
	gob.Register(uuid.UUID{})
	return &StateManager{
		MSync:  new(sync.RWMutex),
		States: make(map[uuid.UUID]*State),
		Store:  sessionStore,
	}
}

func (m *StateManager) SessionWrapper(cb func(http.ResponseWriter, *http.Request, *State)) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		session, err := m.Store.Get(req, "build-storage-session")
		if err == nil {
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
		} else {
			session, err = m.Store.New(req, "build-storage-session")
			if err != nil {
				http.Error(rw, "500 Internal Server Error: Session Malfunction", http.StatusInternalServerError)
				return
			}
		}
		u := NewState()
		m.MSync.Lock()
		m.States[u.Uuid] = u
		m.MSync.Unlock()
		session.Values["session-key"] = u.Uuid
		err = session.Save(req, rw)
		if err != nil {
			log.Println("Failed to save session:", err)
			http.Error(rw, "500 Internal Server Error: Failed to save session", http.StatusInternalServerError)
			return
		}
		cb(rw, req, u)
	}
}
