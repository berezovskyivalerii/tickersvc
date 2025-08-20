package httpctrl

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	ldom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
)

type itemDTO struct {
	SpotSymbol   string `json:"SpotSymbol"`
	FutureSymbol string `json:"FutureSymbol"`
}

type itemsResp struct {
	Items []itemDTO `json:"items"`
}

type PublicListsController struct {
	Q ldom.QueryRepo
}

func NewPublicListsController(q ldom.QueryRepo) *PublicListsController {
	return &PublicListsController{Q: q}
}

func (ctl *PublicListsController) Register(r *gin.Engine) {
	api := r.Group("/api")
	api.GET("/lists/:slug", ctl.bySlug)                   // JSON или text (as_text=1)
	api.GET("/lists", ctl.byTarget)                       // уже умел text
	api.GET("/segments/:source/:seg", ctl.segmentForward) // без редиректа
}

func wantText(ctx *gin.Context) bool {
	v := strings.ToLower(strings.TrimSpace(ctx.Query("as_text")))
	return v == "1" || v == "true" || v == "yes"
}

func (ctl *PublicListsController) bySlug(c *gin.Context) {
	slug := c.Param("slug")
	rows, err := ctl.Q.GetRowsBySlug(c, slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if wantText(c) {
		writeTextRows(c, rows)
		return
	}

	items := make([]itemDTO, 0, len(rows))
	for _, r := range rows {
		fs := "none"
		if r.Futures != nil && *r.Futures != "" {
			fs = *r.Futures
		}
		items = append(items, itemDTO{SpotSymbol: r.Spot, FutureSymbol: fs})
	}
	c.JSON(http.StatusOK, itemsResp{Items: items})
}

func (ctl *PublicListsController) byTarget(ctx *gin.Context) {
	target := strings.TrimSpace(ctx.Query("target"))
	if target == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing target"})
		return
	}
	data, err := ctl.Q.GetTextByTarget(ctx, target) // map[source][]lines
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

// внутренний форвард без 307
func (ctl *PublicListsController) segmentForward(c *gin.Context) {
	source := strings.ToLower(strings.TrimSpace(c.Param("source")))
	seg := strings.TrimSpace(c.Param("seg"))

	switch source {
	case "binance", "bybit", "okx":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "source must be one of: binance, bybit, okx"})
		return
	}
	switch seg {
	case "0", "1", "2", "3", "4":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "seg must be 0, 1, 2, 3, or 4"})
		return
	}

	// подложим slug и переиспользуем bySlug (сохранит ?as_text=1)
	slug := source + "_seg" + seg
	c.Params = append(c.Params, gin.Param{Key: "slug", Value: slug})
	ctl.bySlug(c)
}

func writeTextRows(c *gin.Context, rows []ldom.Row) {
	lines := make([]string, 0, len(rows))
	for _, r := range rows {
		fs := "none"
		if r.Futures != nil && *r.Futures != "" {
			fs = *r.Futures
		}
		lines = append(lines, r.Spot+", "+fs)
	}
	sort.Strings(lines)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	for i, line := range lines {
		if i > 0 {
			_, _ = c.Writer.WriteString("\n")
		}
		_, _ = c.Writer.WriteString(line)
	}
	_, _ = c.Writer.WriteString("\n")
}
