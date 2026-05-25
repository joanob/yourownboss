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
	RevokedAt          *time.Time // siempre nil en entradas activas; conservado por compatibilidad
	CreatedAt          time.Time
}

// SessionCache almacena sesiones de usuario con TTL.
// PERF-05: mantiene un índice inverso userID→[]sessionID para que ClearCompanyID
// sea O(k) en lugar de O(N) sobre todas las sesiones.
type SessionCache struct {
	sessions     map[string]SessionData // sessionID → SessionData
	userSessions map[string][]string    // userID → []sessionID (índice inverso)
	mu           sync.RWMutex
}

// NewSessionCache crea una nueva instancia del cache de sesiones
func NewSessionCache() *SessionCache {
	return &SessionCache{
		sessions:     make(map[string]SessionData),
		userSessions: make(map[string][]string),
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
	sc.userSessions[data.UserID] = append(sc.userSessions[data.UserID], sessionID)
}

// Get obtiene una sesión por ID
// Retorna la sesión y true si existe y no ha expirado, false si no existe o expiró.
func (sc *SessionCache) Get(sessionID string) (SessionData, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	session, exists := sc.sessions[sessionID]
	if !exists {
		return SessionData{}, false
	}

	// Verificar que no ha expirado
	if time.Now().After(session.ExpiresAt) {
		return SessionData{}, false
	}

	return session, true
}

// Revoke elimina inmediatamente una sesión del cache.
// PERF-08: la eliminación inmediata evita que sesiones revocadas acumulen memoria
// hasta el siguiente ciclo de limpieza.
func (sc *SessionCache) Revoke(sessionID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.deleteSession(sessionID)
}

// Delete elimina una sesión del cache.
func (sc *SessionCache) Delete(sessionID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.deleteSession(sessionID)
}

// deleteSession elimina la sesión de ambos mapas. Debe llamarse con el lock adquirido.
func (sc *SessionCache) deleteSession(sessionID string) {
	session, exists := sc.sessions[sessionID]
	if !exists {
		return
	}

	// Eliminar del índice inverso
	userID := session.UserID
	ids := sc.userSessions[userID]
	for i, id := range ids {
		if id == sessionID {
			sc.userSessions[userID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	if len(sc.userSessions[userID]) == 0 {
		delete(sc.userSessions, userID)
	}

	delete(sc.sessions, sessionID)
}

// CleanupExpired elimina sesiones expiradas del cache.
// Se debe llamar periódicamente (cada 6 horas según especificación).
// Con PERF-08, las sesiones revocadas ya se eliminan al momento de la revocación.
func (sc *SessionCache) CleanupExpired() int {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	deleted := 0
	now := time.Now()

	for sessionID, session := range sc.sessions {
		if now.After(session.ExpiresAt) {
			sc.deleteSession(sessionID)
			deleted++
		}
	}

	return deleted
}

// ClearCompanyID pone company_id a nil en todas las sesiones activas de un usuario.
// PERF-05: usa el índice inverso userSessions para ser O(k) en lugar de O(N).
func (sc *SessionCache) ClearCompanyID(userID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for _, sessionID := range sc.userSessions[userID] {
		session, exists := sc.sessions[sessionID]
		if !exists {
			continue
		}
		session.CompanyID = nil
		sc.sessions[sessionID] = session
	}
}

// Count devuelve el número de sesiones en el cache
func (sc *SessionCache) Count() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return len(sc.sessions)
}
