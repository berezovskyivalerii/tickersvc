package httpctrl

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	listsuc "github.com/berezovskyivalerii/tickersvc/internal/usecase/lists"
)

type UpdateController struct {
	UC *listsuc.Updater
}

func NewUpdateController(uc *listsuc.Updater) *UpdateController { return &UpdateController{UC: uc} }

func (c *UpdateController) Register(r *gin.Engine) {
	r.POST("/update", c.update)
}

func (c *UpdateController) update(ctx *gin.Context) {
	var src, tgt *string

	// параметры можно передать query: ?source=okx,bybit&target=upbit
	if q := strings.TrimSpace(ctx.Query("source")); q != "" {
		// поддержим только один source (как в ТЗ), но позволим запятую выбрать первый
		s := strings.Split(q, ",")[0]
		src = &s
	}
	if q := strings.TrimSpace(ctx.Query("target")); q != "" {
		t := strings.Split(q, ",")[0]
		tgt = &t
	}

	res, err := c.UC.Update(ctx, src, tgt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"updated": res})
}
