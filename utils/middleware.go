package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func CheckLoginIn() gin.HandlerFunc {
	return func(context *gin.Context) {

		token:=context.DefaultQuery("token","invalid")

		if token != Token {
			if strings.ToUpper(context.Request.Method) == "GET" {
				context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			} else if strings.ToUpper(context.Request.Method) == "POST" {
				context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			} else {
				context.AbortWithStatus(http.StatusUnauthorized)
			}
		} else {
			context.Next()
		}
	}
}
