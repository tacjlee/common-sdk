package fxmodels

import "github.com/golang-jwt/jwt/v4"

type RealmAccess struct {
	Roles []string `json:"roles"`
}

type TokenClaims struct {
	Email          string      `json:"email"`
	RealmAccess    RealmAccess `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
	jwt.RegisteredClaims
}
