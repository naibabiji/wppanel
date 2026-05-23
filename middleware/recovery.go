package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}
