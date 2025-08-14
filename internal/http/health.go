package http

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/berezovskyivalerii/tickersvc/internal/config"
	"github.com/berezovskyivalerii/tickersvc/internal/store"
)

type HealthHandler struct {
	db    *sql.DB
	build config.BuildInfo
}

func NewHealthHandler(db *sql.DB, build config.BuildInfo) *HealthHandler {
	return &HealthHandler{db: db, build: build}
}

type healthResp struct {
	Status    string            `json:"status"`
	Version   string            `json:"version,omitempty"`
	Commit    string            `json:"commit,omitempty"`
	BuildTime string            `json:"buildTime,omitempty"`
	Uptime    string            `json:"uptime,omitempty"`
	Checks    map[string]string `json:"checks"`
	Now       string            `json:"now,omitempty"`
}

func (h *HealthHandler) Register(r *gin.Engine) {
	r.GET("/health", h.get)
	r.HEAD("/health", h.head)
}

func (h *HealthHandler) head(c *gin.Context) {
	// HEAD должен повторять код из GET /health, но без тела.
	code := h.statusCode()
	c.Status(code)
}

func (h *HealthHandler) get(c *gin.Context) {
	checks := map[string]string{
		"db": "ok",
	}

	// Быстрый ping к БД с маленьким таймаутом.
	if err := store.PingCtx(h.db, 500*time.Millisecond); err != nil {
		checks["db"] = "down"
		resp := healthResp{
			Status: "degraded",
			Checks: checks,
		}
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}

	uptime := time.Since(h.build.StartedAt).Truncate(time.Second).String()
	resp := healthResp{
		Status:    "ok",
		Version:   h.build.Version,
		Commit:    h.build.Commit,
		BuildTime: h.build.BuildTime,
		Uptime:    uptime,
		Checks:    checks,
		Now:       time.Now().UTC().Format(time.RFC3339),
	}
	c.JSON(http.StatusOK, resp)
}

// statusCode используется для HEAD: повторить логику, но без двойного пинга.
func (h *HealthHandler) statusCode() int {
	if err := store.PingCtx(h.db, 500*time.Millisecond); err != nil {
		return http.StatusServiceUnavailable
	}
	return http.StatusOK
}
