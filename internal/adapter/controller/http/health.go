package httpctrl

import (
	"github.com/gin-gonic/gin"

	presenter "github.com/berezovskyivalerii/tickersvc/internal/adapter/presenter/health"
	usecase "github.com/berezovskyivalerii/tickersvc/internal/usecase/health"
)

type ReadinessRunner struct{ UC *usecase.ReadinessInteractor }

func (r ReadinessRunner) Execute(c *gin.Context, in usecase.ReadinessInput) usecase.ReadinessOutput {
	return r.UC.Execute(c.Request.Context(), in)
}

type HealthController struct {
	run ReadinessRunner
}

func NewHealthController(run ReadinessRunner) *HealthController {
	return &HealthController{run: run}
}

func (h *HealthController) Register(r *gin.Engine) {
	r.GET("/health", h.get)
	r.HEAD("/health", h.head)
}

func (h *HealthController) get(c *gin.Context) {
	out := h.run.Execute(c, usecase.ReadinessInput{})
	code, body := presenter.Map(out)
	c.JSON(code, body)
}

func (h *HealthController) head(c *gin.Context) {
	out := h.run.Execute(c, usecase.ReadinessInput{})
	code, _ := presenter.Map(out)
	c.Status(code)
}
