package adminauth

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Middleware struct {
	apiKey string
}

func NewFromEnv() *Middleware {
	return &Middleware{
		apiKey: strings.TrimSpace(os.Getenv("ADMIN_API_KEY")),
	}
}

func (m *Middleware) checkKey(r *http.Request) bool {
	if m.apiKey == "" {
		return false
	}
	if k := strings.TrimSpace(r.Header.Get("X-API-Key")); k != "" {
		return k == m.apiKey
	}
	const pfx = "Bearer "
	if auth := strings.TrimSpace(r.Header.Get("Authorization")); strings.HasPrefix(auth, pfx) {
		return strings.TrimSpace(auth[len(pfx):]) == m.apiKey
	}
	return false
}

func (m *Middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.apiKey == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				gin.H{"error": "server not configured (ADMIN_API_KEY is empty)"})
			return
		}
		if !m.checkKey(c.Request) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
