package fxmodel

import "github.com/golang-jwt/jwt/v4"

// RealmAccess for Keycloak tokens
type RealmAccess struct {
	Roles []string `json:"roles"`
}

// TokenClaims for Keycloak tokens
type TokenClaims struct {
	Email          string      `json:"email"`
	RealmAccess    RealmAccess `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
	jwt.RegisteredClaims
}

// JWTClaims for gauth-service tokens
type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
