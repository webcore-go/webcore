package session

import "github.com/webcore-go/webcore/port/auth"

type AuthSession struct {
	Backend auth.ISessionStore
}

func (s *AuthSession) SetSessionStore(backend auth.ISessionStore) {
	s.Backend = backend
}

func (s *AuthSession) Install(args ...any) error {
	// Tidak melakukan apa-apa
	return nil
}

func (s *AuthSession) Uninstall() error {
	// Tidak melakukan apa-apa
	return nil
}

func (s *AuthSession) GetSessionStore() auth.ISessionStore {
	return s.Backend
}
