package cache

import (
	"sync"
	"time"
)

// SessionData representa los datos de una sesión en cache
type SessionData struct {
	SessionID          string
	UserID             string
	CompanyID          *string // nullable si usuario no tiene empresa
	VerificationString string
	TokenHash          string // hash del refresh token
	ExpiresAt          time.Time
	RevokedAt          *time.Time // nil si sesión no está revocada
	CreatedAt          time.Time
}

// SessionCache almacena sesiones de usuario con TTL
type SessionCache struct {
	sessions map[string]SessionData
	mu       sync.RWMutex
}

// NewSessionCache crea una nueva instancia del cache de sesiones
func NewSessionCache() *SessionCache {
	return &SessionCache{
		sessions: make(map[string]SessionData),
	}
}

// Set almacena una sesión en el cache (alias de Store para compatibilidad)
func (sc *SessionCache) Set(sessionID string, session SessionData) {
	sc.Store(sessionID, session)
}

// Store almacena una sesión en el cache (interfaz SessionCache del auth/service)
func (sc *SessionCache) Store(sessionID string, data SessionData) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.sessions[sessionID] = data
}

// Get obtiene una sesión por ID
// Retorna la sesión y true si existe y no ha expirado, false si no existe, expiró o está revocada
func (sc *SessionCache) Get(sessionID string) (SessionData, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	session, exists := sc.sessions[sessionID]
	if !exists {
		return SessionData{}, false
	}

	// Verificar que no está revocada
	if session.RevokedAt != nil {
		return SessionData{}, false
	}

	// Verificar que no ha expirado
	if time.Now().After(session.ExpiresAt) {
		return SessionData{}, false
	}

	return session, true
}

// Revoke marca una sesión como revocada
func (sc *SessionCache) Revoke(sessionID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	session, exists := sc.sessions[sessionID]
	if exists {
		now := time.Now()
		session.RevokedAt = &now
		sc.sessions[sessionID] = session
	}
}

// Delete elimina una sesión del cache
func (sc *SessionCache) Delete(sessionID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	delete(sc.sessions, sessionID)
}

// CleanupExpired elimina sesiones expiradas y revocadas del cache
// Se debe llamar periódicamente (cada 6 horas según especificación)
func (sc *SessionCache) CleanupExpired() int {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	deleted := 0
	now := time.Now()

	for sessionID, session := range sc.sessions {
		// Eliminar si expiró o está revocada
		if now.After(session.ExpiresAt) || session.RevokedAt != nil {
			delete(sc.sessions, sessionID)
			deleted++
		}
	}

	return deleted
}

// ClearCompanyID sets company_id to nil for all active sessions of a given user.
// Called when a user's company is deleted so that renewed tokens reflect the change.
func (sc *SessionCache) ClearCompanyID(userID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for sessionID, session := range sc.sessions {
		if session.UserID == userID && session.RevokedAt == nil {
			session.CompanyID = nil
			sc.sessions[sessionID] = session
		}
	}
}

// Count devuelve el número de sesiones en el cache
func (sc *SessionCache) Count() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return len(sc.sessions)
}
