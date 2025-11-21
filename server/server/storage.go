package server

import "sync"

type SessionStorage struct {
	mu       sync.RWMutex
	sessions map[int64]*GameSession
}

func MakeSessionStorage() *SessionStorage {
	return &SessionStorage{
		sessions: make(map[int64]*GameSession),
	}
}

func (s *SessionStorage) Add(session *GameSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.sessionId] = session
}

func (s *SessionStorage) Get(sessionId int64) (*GameSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[sessionId]
	return session, ok
}

func (s *SessionStorage) GetAll() []*GameSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	slice := make([]*GameSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		slice = append(slice, session)
	}
	return slice
}

func (s *SessionStorage) Delete(sessionId int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionId)
}
