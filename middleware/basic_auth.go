package middleware

import (
	"net/http"

	"github.com/naibabiji/wp-panel/config"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type BasicAuthChecker struct {
	RecordAttempt func(ip string, attemptType string)
	IsBanned      func(ip string) bool
}

func BasicAuth(checker *BasicAuthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if checker.IsBanned != nil && checker.IsBanned(ip) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "IP已被临时封禁，请稍后再试",
			})
			return
		}

		user, pass, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="WP Panel Authentication"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		cfg := config.AppConfig
		if cfg == nil {
			c.Header("WWW-Authenticate", `Basic realm="WP Panel Authentication"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if user != cfg.BasicAuth.Username ||
			bcrypt.CompareHashAndPassword([]byte(cfg.BasicAuth.PasswordHash), []byte(pass)) != nil {
			if checker.RecordAttempt != nil {
				checker.RecordAttempt(ip, "basic_auth")
			}
			c.Header("WWW-Authenticate", `Basic realm="WP Panel Authentication"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("authenticated_user", user)
		c.Next()
	}
}
