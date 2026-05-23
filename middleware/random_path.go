package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RandomPath(suffix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := strings.Trim(c.Request.URL.Path, "/")
		if path == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		parts := strings.SplitN(path, "/", 2)
		if len(parts) == 0 || parts[0] != suffix {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Next()
	}
}
