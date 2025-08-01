package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Athooh/social-network/pkg/session"
)

// JWT configuration
type JWTConfig struct {
	SecretKey     string
	TokenDuration time.Duration
	Issuer        string
}

// StandardClaims represents the standard JWT claims
type StandardClaims struct {
	ExpiresAt int64  `json:"exp,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	NotBefore int64  `json:"nbf,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	Subject   string `json:"sub,omitempty"`
}

// Claims represents the JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	StandardClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID string, config JWTConfig) (string, error) {
	// Create the claims
	claims := Claims{
		UserID: userID,
		StandardClaims: StandardClaims{
			ExpiresAt: time.Now().Add(config.TokenDuration).Unix(),
			IssuedAt:  time.Now().Unix(),
			NotBefore: time.Now().Unix(),
			Issuer:    config.Issuer,
			Subject:   userID,
		},
	}

	// Create the JWT header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	// Marshal header and claims to JSON
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	// Base64 encode header and claims
	headerBase64 := base64URLEncode(headerJSON)
	claimsBase64 := base64URLEncode(claimsJSON)

	// Create the signature
	signatureInput := headerBase64 + "." + claimsBase64
	signature := createSignature(signatureInput, config.SecretKey)

	// Combine all parts to form the JWT token
	token := signatureInput + "." + signature

	return token, nil
}

// ValidateToken validates a JWT token and optionally checks for a valid session
func ValidateToken(tokenString string, config JWTConfig, sessionStore ...session.SessionStore) (*Claims, error) {
	// Split the token into parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerBase64, claimsBase64, signatureBase64 := parts[0], parts[1], parts[2]

	// Verify the signature
	signatureInput := headerBase64 + "." + claimsBase64
	expectedSignature := createSignature(signatureInput, config.SecretKey)

	if signatureBase64 != expectedSignature {
		return nil, errors.New("invalid token signature")
	}

	// Decode the claims
	claimsJSON, err := base64URLDecode(claimsBase64)
	if err != nil {
		return nil, err
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, err
	}

	// Validate expiration
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("token has expired")
	}

	// Validate not before
	if claims.NotBefore > time.Now().Unix() {
		return nil, errors.New("token not valid yet")
	}

	// If a session store is provided, validate that there's an active session for this user
	if len(sessionStore) > 0 && sessionStore[0] != nil {
		// Get all sessions for the user
		sessions, err := sessionStore[0].GetUserSessions(claims.UserID)
		if err != nil {
			return nil, errors.New("failed to validate user session")
		}

		// Check if there's at least one valid session
		now := time.Now()
		validSessionFound := false
		for _, session := range sessions {
			if session.ExpiresAt.After(now) {
				validSessionFound = true
				break
			}
		}

		if !validSessionFound {
			return nil, errors.New("no valid session found for user")
		}
	}

	return &claims, nil
}

// ExtractTokenFromRequest extracts the JWT token from the request
func ExtractTokenFromRequest(r *http.Request) (string, error) {
	// First try to get from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1], nil
		}
	}

	// If not found in header, try to get from URL query parameter
	token := r.URL.Query().Get("token")
	if token != "" {
		return token, nil
	}

	// If token not found in either location
	return "", errors.New("token not found in Authorization header or URL parameter")
}

// Helper functions for JWT operations

// base64URLEncode encodes data to base64url format
func base64URLEncode(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	encoded = strings.ReplaceAll(encoded, "+", "-")
	encoded = strings.ReplaceAll(encoded, "/", "_")
	encoded = strings.TrimRight(encoded, "=")
	return encoded
}

// base64URLDecode decodes base64url format to bytes
func base64URLDecode(str string) ([]byte, error) {
	// Add padding if needed
	padding := 4 - (len(str) % 4)
	if padding < 4 {
		str += strings.Repeat("=", padding)
	}

	// Replace URL-safe characters
	str = strings.ReplaceAll(str, "-", "+")
	str = strings.ReplaceAll(str, "_", "/")

	return base64.StdEncoding.DecodeString(str)
}

// createSignature creates an HMAC-SHA256 signature for the JWT
func createSignature(input, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(input))
	return base64URLEncode(h.Sum(nil))
}
