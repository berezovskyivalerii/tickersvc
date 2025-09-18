package httpctrl

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeUpdater struct {
	lastSrc *string
	lastTgt *string
	res     map[string]int
	err     error
}

func (u *fakeUpdater) Update(ctx context.Context, src, tgt *string) (map[string]int, error) {
	u.lastSrc, u.lastTgt = src, tgt
	return u.res, u.err
}

func strPtr(s string) *string { return &s }

func TestUpdate_All_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upd := &fakeUpdater{res: map[string]int{"okx_to_upbit": 2, "binance_to_upbit": 5}}
	r := gin.New()
	// Имитация контроллера: без параметров → src=nil,tgt=nil
	r.POST("/update", func(c *gin.Context) {
		res, err := upd.Update(c, nil, nil)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"updated": res})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/update", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var got struct{ Updated map[string]int `json:"updated"` }
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("json: %v", err)
	}
	if got.Updated["okx_to_upbit"] != 2 || got.Updated["binance_to_upbit"] != 5 {
		t.Fatalf("payload: %#v", got)
	}
	if upd.lastSrc != nil || upd.lastTgt != nil {
		t.Fatalf("src/tgt should be nil")
	}
}

func TestUpdate_WithParams_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upd := &fakeUpdater{res: map[string]int{"okx_to_upbit": 2}}
	r := gin.New()
	// Имитация контроллера: берём первый из списка через запятую и обрезаем пробелы
	r.POST("/update", func(c *gin.Context) {
		var src, tgt *string
		if q := c.Query("source"); q != "" {
			val := q
			// first, trim
			if i := indexRune(q, ','); i >= 0 {
				val = q[:i]
			}
			s := trim(val)
			if s != "" {
				src = &s
			}
		}
		if q := c.Query("target"); q != "" {
			val := q
			if i := indexRune(q, ','); i >= 0 {
				val = q[:i]
			}
			t := trim(val)
			if t != "" {
				tgt = &t
			}
		}
		res, err := upd.Update(c, src, tgt)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"updated": res})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/update?source=%20okx,%20bybit&target=%20upbit,coinbase", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if upd.lastSrc == nil || *upd.lastSrc != "okx" || upd.lastTgt == nil || *upd.lastTgt != "upbit" {
		t.Fatalf("params not passed: src=%v tgt=%v", upd.lastSrc, upd.lastTgt)
	}
}

func TestUpdate_Error_BubblesUp(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upd := &fakeUpdater{err: errors.New("boom")}
	r := gin.New()
	r.POST("/update", func(c *gin.Context) {
		_, err := upd.Update(c, nil, nil)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/update", nil)
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Fatalf("want 500, got %d body=%s", w.Code, w.Body.String())
	}
	var got map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if got["error"] != "boom" {
		t.Fatalf("error payload: %#v", got)
	}
}

// helpers (простые аналоги strings функций, чтобы не тащить extra импорт)
func indexRune(s string, r rune) int {
	for i, ch := range s {
		if ch == r {
			return i
		}
	}
	return -1
}
func trim(s string) string {
	// только пробелы по краям — для теста достаточно
	for len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}
	for len(s) > 0 && s[len(s)-1] == ' ' {
		s = s[:len(s)-1]
	}
	return s
}
