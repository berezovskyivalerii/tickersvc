package adminauth

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Middleware struct {
	apiKey string
	nets   []*net.IPNet
	requireBoth bool
}

func NewFromEnv() *Middleware {
	m := &Middleware{
		apiKey: strings.TrimSpace(os.Getenv("ADMIN_API_KEY")),
	}
	if v := strings.TrimSpace(os.Getenv("ADMIN_REQUIRE_BOTH")); v == "1" || strings.EqualFold(v, "true") {
		m.requireBoth = true
	}
	if cidrs := strings.TrimSpace(os.Getenv("ADMIN_TRUSTED_CIDRS")); cidrs != "" {
		for _, c := range strings.Split(cidrs, ",") {
			c = strings.TrimSpace(c)
			if c == "" {
				continue
			}
			if _, ipn, err := net.ParseCIDR(c); err == nil {
				m.nets = append(m.nets, ipn)
			}
		}
	}
	return m
}

func (m *Middleware) allowedIP(ipStr string) bool {
	if len(m.nets) == 0 {
		return false
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, n := range m.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func (m *Middleware) checkKey(r *http.Request) bool {
	if m.apiKey == "" {
		return false
	}
	if k := strings.TrimSpace(r.Header.Get("X-API-Key")); k != "" {
		return k == m.apiKey
	}
	const p = "Bearer "
	if auth := strings.TrimSpace(r.Header.Get("Authorization")); strings.HasPrefix(auth, p) {
		return strings.TrimSpace(auth[len(p):]) == m.apiKey
	}
	return false
}

func (m *Middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		ipOK := m.allowedIP(clientIP)
		keyOK := m.checkKey(c.Request)

		allow := false
		if m.requireBoth {
			allow = ipOK && keyOK
		} else {
			allow = ipOK || keyOK
		}

		if !allow {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "admin endpoint restricted",
			})
			return
		}
		c.Next()
	}
}
