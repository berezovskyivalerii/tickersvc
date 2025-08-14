package httpctrl

import (
	"net/http"

	"github.com/gin-gonic/gin"

	ldom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
)

type ListsController struct {
	Q ldom.QueryRepo
}

func NewListsController(q ldom.QueryRepo) *ListsController { return &ListsController{Q: q} }

func (c *ListsController) Register(r *gin.Engine) {
	r.GET("/lists", c.all)
	r.GET("/lists/:target", c.byTarget)
	r.GET("/list", c.single) // ?source=okx&target=upbit → text/plain
}

func (c *ListsController) all(ctx *gin.Context) {
	out, err := c.Q.GetAllText(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, out)
}

func (c *ListsController) byTarget(ctx *gin.Context) {
	tgt := ctx.Param("target")
	out, err := c.Q.GetTextByTarget(ctx, tgt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, out)
}

func (c *ListsController) single(ctx *gin.Context) {
	src := ctx.Query("source")
	tgt := ctx.Query("target")
	if src == "" || tgt == "" {
		ctx.String(http.StatusBadRequest, "missing source or target")
		return
	}
	// slug по схеме "<src>_to_<tgt>"
	slug := src + "_to_" + tgt
	lines, err := c.Q.GetTextBySlug(ctx, slug)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	// строго «список строк "тикер, тикер"»
	for i, s := range lines {
		if i > 0 {
			ctx.Writer.WriteString("\n")
		}
		ctx.Writer.WriteString(s)
	}
	ctx.Writer.WriteString("\n")
	ctx.Header("Content-Type", "text/plain; charset=utf-8")
}
