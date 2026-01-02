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
func GetUserID(ctx *gin.Context) int {
	if val, exists := ctx.Get("userID"); exists {
		if userID, ok := val.(int); ok {
			return userID
		}
	}
	return 0
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

// GetCompanyUUID retrieves the company UUID from the Gin context
func GetCompanyUUID(ctx *gin.Context) string {
	if val, exists := ctx.Get("companyUUID"); exists {
		if companyUUID, ok := val.(string); ok {
			return companyUUID
		}
	}
	return ""
}

// GetCompanyID retrieves the company ID from the Gin context
func GetCompanyID(ctx *gin.Context) int {
	if val, exists := ctx.Get("companyID"); exists {
		if companyID, ok := val.(int); ok {
			return companyID
		}
	}
	return 0
}
