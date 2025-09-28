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
