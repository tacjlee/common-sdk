package fxutil

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tacjlee/common-sdk/packages/fxmodel"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

func CreateToken(ttl time.Duration, payload interface{}, privateKey string) (string, error) {
	decodedPrivateKey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("could not decode key: %w", err)
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(decodedPrivateKey)

	if err != nil {
		return "", fmt.Errorf("create: parse key: %w", err)
	}

	now := time.Now().UTC()

	claims := make(jwt.MapClaims)
	claims["sub"] = payload
	claims["exp"] = now.Add(ttl).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()

	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)

	if err != nil {
		return "", fmt.Errorf("create: sign token: %w", err)
	}

	return token, nil
}

func ValidateToken(token string, publicKey string) (interface{}, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("could not decode: %w", err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(decodedPublicKey)

	if err != nil {
		return "", fmt.Errorf("validate: parse key: %w", err)
	}

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %s", t.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("validate: invalid token")
	}

	return claims["sub"], nil
}

// ExtractTokenPayload returns the base64-decoded JSON payload segment of a JWT.
func ExtractTokenPayload(jwtToken string) ([]byte, error) {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT token format")
	}
	payload, err := jwt.DecodeSegment(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode token payload: %v", err)
	}
	return payload, nil
}

func ExtractKeycloakRoles(keycloakToken string) ([]string, error) {
	payload, err := ExtractTokenPayload(keycloakToken)
	if err != nil {
		return nil, err
	}
	var claims fxmodel.TokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token claims: %v", err)
	}
	roles := claims.RealmAccess.Roles
	for _, resource := range claims.ResourceAccess {
		roles = append(roles, resource.Roles...)
	}
	return roles, nil
}

func ExtractTokenClaims[T any](jwtToken string) (T, error) {
	var claims T
	payload, err := ExtractTokenPayload(jwtToken)
	if err != nil {
		return claims, err
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return claims, fmt.Errorf("failed to unmarshal token claims: %v", err)
	}
	return claims, nil
}

// ExtractClaim returns the JWT claim at the given dotted path (e.g. "ext.email",
// "realm_access.roles"). Returns (nil, nil) if the path is missing.
func ExtractClaim(jwtToken, path string) (any, error) {
	payload, err := ExtractTokenPayload(jwtToken)
	if err != nil {
		return nil, err
	}
	var root map[string]any
	if err := json.Unmarshal(payload, &root); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token claims: %v", err)
	}
	return lookupPath(root, path), nil
}

// ExtractClaimString returns the first non-empty string claim across the given
// paths. Intended for tokens that may carry the same value in different shapes,
// e.g. ExtractClaimString(tok, "email", "ext.email").
func ExtractClaimString(jwtToken string, paths ...string) (string, error) {
	payload, err := ExtractTokenPayload(jwtToken)
	if err != nil {
		return "", err
	}
	var root map[string]any
	if err := json.Unmarshal(payload, &root); err != nil {
		return "", fmt.Errorf("failed to unmarshal token claims: %v", err)
	}
	for _, p := range paths {
		if s, ok := lookupPath(root, p).(string); ok && s != "" {
			return s, nil
		}
	}
	return "", nil
}

// ExtractClaimStrings merges string-array (or single-string) claims across the
// given paths. Intended for roles, e.g.
// ExtractClaimStrings(tok, "realm_access.roles", "ext.role").
func ExtractClaimStrings(jwtToken string, paths ...string) ([]string, error) {
	payload, err := ExtractTokenPayload(jwtToken)
	if err != nil {
		return nil, err
	}
	var root map[string]any
	if err := json.Unmarshal(payload, &root); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token claims: %v", err)
	}
	var out []string
	for _, p := range paths {
		switch v := lookupPath(root, p).(type) {
		case string:
			if v != "" {
				out = append(out, v)
			}
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok && s != "" {
					out = append(out, s)
				}
			}
		}
	}
	return out, nil
}

// lookupPath walks a dotted path over a decoded JSON tree. Returns nil for
// missing segments or when a non-map value is encountered mid-path.
func lookupPath(root any, path string) any {
	if path == "" {
		return root
	}
	cur := root
	for _, seg := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = m[seg]
		if !ok {
			return nil
		}
	}
	return cur
}

func QueryInt(ctx *gin.Context, key string, defaultValue int) int {
	valStr := ctx.Query(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}
