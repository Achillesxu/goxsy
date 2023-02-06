// Package middlewares
// Time    : 2023/2/2 16:47
// Author  : xushiyin
// contact : yuqingxushiyin@gmail.com
package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"goxsy/ginAuth/handlers"
	"net/http"
	"os"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenVal := c.GetHeader("authorization")
		claims := &handlers.Claims{}
		tkn, err := jwt.ParseWithClaims(tokenVal, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		if tkn == nil || !tkn.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		c.Next()
	}
}
