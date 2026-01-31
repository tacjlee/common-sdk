package fxcontext

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func DefaultQueryInt(ctx *gin.Context, key string, defaultValue int) int {
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

// GetUserID retrieves the user ID from the Gin context
func GetUserID(ctx *gin.Context) string {
	if val, exists := ctx.Get("userID"); exists {
		if userID, ok := val.(string); ok {
			return userID
		}
	}
	return ""
}

// GetEmail retrieves the email from the Gin context
func GetEmail(ctx *gin.Context) string {
	if val, exists := ctx.Get("email"); exists {
		if email, ok := val.(string); ok {
			return email
		}
	}
	return ""
}

// GetRole retrieves the role from the Gin context
func GetRole(ctx *gin.Context) string {
	if val, exists := ctx.Get("role"); exists {
		if role, ok := val.(string); ok {
			return role
		}
	}
	return ""
}

// GetUsername retrieves the username from the Gin context
func GetUsername(ctx *gin.Context) string {
	if val, exists := ctx.Get("username"); exists {
		if username, ok := val.(string); ok {
			return username
		}
	}
	return ""
}

// GetOrgUnitID retrieves the organization unit ID from the Gin context
func GetOrgUnitID(ctx *gin.Context) string {
	if val, exists := ctx.Get("orgUnitID"); exists {
		if orgUnitID, ok := val.(string); ok {
			return orgUnitID
		}
	}
	return ""
}

// GetTenantID retrieves the tenant ID from the Gin context
func GetTenantID(ctx *gin.Context) string {
	if val, exists := ctx.Get("tenantID"); exists {
		if orgUnitID, ok := val.(string); ok {
			return orgUnitID
		}
	}
	return ""
}
