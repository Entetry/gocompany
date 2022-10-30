package model

import "github.com/golang-jwt/jwt"

// RefreshSession refresh token struct
type RefreshSession struct {
	RefreshToken string
	UserID       string
	ExpiresAt    int64
	TokenParam
}

// TokenParam Browser fingerprint struct
type TokenParam struct {
	UserAgent   string
	Fingerprint string
	IP          string
}

// Claim Jwt Claim struct
type Claim struct {
	UserID string
	jwt.StandardClaims
}
