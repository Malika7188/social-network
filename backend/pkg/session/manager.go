package session

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

// SessionManager handles user sessions
type SessionManager struct {
	db            SessionStore
	cookieName    string
	cookieDomain  string
	cookieSecure  bool
	cookieMaxAge  int
	sessionMaxAge time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(store SessionStore, cookieName, cookieDomain string, cookieSecure bool, maxAge int) *SessionManager {
	return &SessionManager{
		db:            store,
		cookieName:    cookieName,
		cookieDomain:  cookieDomain,
		cookieSecure:  cookieSecure,
		cookieMaxAge:  maxAge,
		sessionMaxAge: time.Duration(maxAge) * time.Second,
	}
}

// CreateSession creates a new session for the user
func (sm *SessionManager) CreateSession(w http.ResponseWriter, userID string) error {
	// Always delete any existing sessions for the user
	err := sm.db.DeleteUserSessions(userID)
	if err != nil {
		return err
	}

	// Create a new session
	expiresAt := time.Now().Add(sm.sessionMaxAge)
	sessionID, err := sm.db.CreateSession(userID, expiresAt)
	if err != nil {
		return err
	}

	// Set session cookie
	cookie := &http.Cookie{
		Name:     sm.cookieName,
		Value:    sessionID,
		Path:     "/",
		Domain:   sm.cookieDomain,
		Expires:  expiresAt,
		MaxAge:   sm.cookieMaxAge,
		Secure:   sm.cookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)
	return nil
}

// GetUserFromSession retrieves the user ID from the session
func (sm *SessionManager) GetUserFromSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sm.cookieName)
	if err != nil {
		return "", errors.New("no session cookie found")
	}

	sessionID := cookie.Value
	userID, expiresAt, err := sm.db.GetSession(sessionID)
	if err != nil {
		return "", err
	}

	if time.Now().After(expiresAt) {
		sm.db.DeleteSession(sessionID)
		return "", errors.New("session expired")
	}

	return userID, nil
}

// ClearSession removes the session
func (sm *SessionManager) ClearSession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(sm.cookieName)
	if err == nil {
		sessionID := cookie.Value
		if err := sm.db.DeleteSession(sessionID); err != nil {
			return err
		}
	}

	// Clear the cookie regardless of whether we found a session
	expiredCookie := &http.Cookie{
		Name:     sm.cookieName,
		Value:    "",
		Path:     "/",
		Domain:   sm.cookieDomain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   sm.cookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, expiredCookie)
	return nil
}

// ClearAllUserSessions removes all sessions for a user
func (sm *SessionManager) ClearAllUserSessions(userID string) error {
	return sm.db.DeleteUserSessions(userID)
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPassword compares a bcrypt hashed password with its possible plaintext equivalent
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateUUID generates a new UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GetSessionStore returns the session store
func (sm *SessionManager) GetSessionStore() SessionStore {
	return sm.db
}
