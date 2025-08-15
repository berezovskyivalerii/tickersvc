package httpctrl

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	ldom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
)

type PublicListsController struct {
	Q ldom.QueryRepo
}

func NewPublicListsController(q ldom.QueryRepo) *PublicListsController {
	return &PublicListsController{Q: q}
}

func (c *PublicListsController) Register(r *gin.Engine) {
	api := r.Group("/api")
	api.GET("/lists/:slug", c.bySlug) // GET /api/lists/okx_to_upbit[?as_text=1]
	api.GET("/lists", c.byTarget)     // GET /api/lists?target=upbit[&as_text=1]
}

func wantText(ctx *gin.Context) bool {
	v := strings.ToLower(strings.TrimSpace(ctx.Query("as_text")))
	return v == "1" || v == "true" || v == "yes"
}

func (c *PublicListsController) bySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")
	lines, err := c.Q.GetTextBySlug(ctx, slug)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if wantText(ctx) {
		ctx.Header("Content-Type", "text/plain; charset=utf-8")
		for i, s := range lines {
			if i > 0 {
				_, _ = ctx.Writer.WriteString("\n")
			}
			_, _ = ctx.Writer.WriteString(s)
		}
		_, _ = ctx.Writer.WriteString("\n")
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"slug": slug, "items": lines})
}

func (c *PublicListsController) byTarget(ctx *gin.Context) {
	target := strings.TrimSpace(ctx.Query("target"))
	if target == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing target"})
		return
	}
	data, err := c.Q.GetTextByTarget(ctx, target) // map[source][]lines
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if wantText(ctx) {
		keys := make([]string, 0, len(data))
		for k := range data {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		ctx.Header("Content-Type", "text/plain; charset=utf-8")
		first := true
		for _, k := range keys {
			for _, s := range data[k] {
				if !first {
					_, _ = ctx.Writer.WriteString("\n")
				} else {
					first = false
				}
				_, _ = ctx.Writer.WriteString(s)
			}
		}
		if !first {
			_, _ = ctx.Writer.WriteString("\n")
		}
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"target": target, "sources": data})
}
