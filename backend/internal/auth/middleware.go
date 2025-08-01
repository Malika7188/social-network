package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Athooh/social-network/pkg/httputil"
)

// contextKey is a custom type for context keys
type contextKey string

// UserIDKey is the key for storing the user ID in the request context
const UserIDKey contextKey = "userID"

// RequireAuth now handles both session and JWT authentication automatically
func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for OPTIONS requests
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		var userID string
		var err error

		// Check if it's an Electron client - check both header and query parameter
		clientType := r.Header.Get("X-Client-Type")
		if clientType == "" {
			// Fallback to query parameter for WebSocket connections
			clientType = r.URL.Query().Get("clientType")
		}

		if clientType == "electron" {
			// Use JWT authentication for Electron
			tokenString, tokenErr := ExtractTokenFromRequest(r)
			if tokenErr != nil {
				// For WebSocket, try to get token from query parameter
				if tokenString = r.URL.Query().Get("token"); tokenString == "" {
					httputil.SendError(w, http.StatusUnauthorized, fmt.Sprintf("(ExtractTokenFromRequest) Unauthorized: %s", tokenErr.Error()), true)
					return
				}
			}

			claims, validateErr := ValidateToken(tokenString, s.jwtConfig)
			if validateErr != nil {
				httputil.SendError(w, http.StatusUnauthorized, fmt.Sprintf("(ValidateToken) Unauthorized: %s", validateErr.Error()), true)
				return
			}

			userID = claims.UserID
		} else {
			// Use session authentication for web clients
			userID, err = s.sessionManager.GetUserFromSession(r)
			if err != nil {
				httputil.SendError(w, http.StatusUnauthorized, fmt.Sprintf("(GetUserFromSession) Unauthorized: %s", err.Error()), true)
				return
			}
		}

		// Store user ID in request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// RequireJWTAuth is a middleware that requires JWT authentication
func (s *Service) RequireJWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for OPTIONS requests
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from request
		tokenString, err := ExtractTokenFromRequest(r)
		if err != nil {
			httputil.SendError(w, http.StatusUnauthorized, fmt.Sprintf("(ExtractTokenFromRequest) Unauthorized: %s", err.Error()), true)
			return
		}

		// Validate token
		claims, err := ValidateToken(tokenString, s.jwtConfig)
		if err != nil {
			httputil.SendError(w, http.StatusUnauthorized, fmt.Sprintf("(ValidateToken) Unauthorized: %s", err.Error()), true)
			return
		}

		// Store user ID in request context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
