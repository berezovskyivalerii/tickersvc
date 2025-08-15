package httpctrl_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	httpctrl "github.com/berezovskyivalerii/tickersvc/internal/adapter/controller/http"
	ldom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
)

type fakeQueryRepo struct {
	bySlug    map[string][]string
	byTarget  map[string]map[string][]string
	errSlug   error
	errTarget error
	errAll    error
}

func (f *fakeQueryRepo) GetAllText(ctx context.Context) (map[string]map[string][]string, error) {
	return nil, f.errAll
}
func (f *fakeQueryRepo) GetTextByTarget(ctx context.Context, targetSlug string) (map[string][]string, error) {
	return f.byTarget[targetSlug], f.errTarget
}
func (f *fakeQueryRepo) GetTextBySlug(ctx context.Context, slug string) ([]string, error) {
	return f.bySlug[slug], f.errSlug
}

var _ ldom.QueryRepo = (*fakeQueryRepo)(nil)

func newRouterWithPublic(q ldom.QueryRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	httpctrl.NewPublicListsController(q).Register(r)
	return r
}

func TestPublicLists_BySlug_JSON_OK(t *testing.T) {
	fq := &fakeQueryRepo{
		bySlug: map[string][]string{
			"okx_to_upbit": {"AAA-USDT, AAA-USDT-SWAP", "CCC-USDT, none"},
		},
	}
	r := newRouterWithPublic(fq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/lists/okx_to_upbit", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var got struct {
		Slug  string   `json:"slug"`
		Items []string `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("json: %v", err)
	}
	if got.Slug != "okx_to_upbit" || len(got.Items) != 2 {
		t.Fatalf("payload: %#v", got)
	}
}

func TestPublicLists_BySlug_Text_OK(t *testing.T) {
	fq := &fakeQueryRepo{
		bySlug: map[string][]string{
			"okx_to_upbit": {"AAA-USDT, AAA-USDT-SWAP", "CCC-USDT, none"},
		},
	}
	r := newRouterWithPublic(fq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/lists/okx_to_upbit?as_text=1", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("content-type=%s", ct)
	}
	want := "AAA-USDT, AAA-USDT-SWAP\nCCC-USDT, none\n"
	if w.Body.String() != want {
		t.Fatalf("body:\n%s\nwant:\n%s", w.Body.String(), want)
	}
}

func TestPublicLists_BySlug_RepoError(t *testing.T) {
	fq := &fakeQueryRepo{errSlug: errors.New("boom")}
	r := newRouterWithPublic(fq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/lists/okx_to_upbit", nil)
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Fatalf("want 500, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestPublicLists_ByTarget_JSON_OK(t *testing.T) {
	fq := &fakeQueryRepo{
		byTarget: map[string]map[string][]string{
			"upbit": {
				"okx":     {"AAA-USDT, AAA-USDT-SWAP"},
				"binance": {"BBB-USDT, none"},
			},
		},
	}
	r := newRouterWithPublic(fq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/lists?target=upbit", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var got struct {
		Target  string              `json:"target"`
		Sources map[string][]string `json:"sources"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("json: %v", err)
	}
	if got.Target != "upbit" || len(got.Sources) != 2 {
		t.Fatalf("payload: %#v", got)
	}
}

func TestPublicLists_ByTarget_Text_OK_SortedSources(t *testing.T) {
	// намеренно источники в другом порядке — контроллер должен отсортировать по ключам
	fq := &fakeQueryRepo{
		byTarget: map[string]map[string][]string{
			"upbit": {
				"okx":     {"O1"},
				"binance": {"B1", "B2"},
			},
		},
	}
	r := newRouterWithPublic(fq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/lists?target=upbit&as_text=true", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("content-type=%s", ct)
	}
	// ожидаем binance (B1,B2) затем okx (O1)
	want := "B1\nB2\nO1\n"
	if w.Body.String() != want {
		t.Fatalf("body:\n%s\nwant:\n%s", w.Body.String(), want)
	}
}

func TestPublicLists_ByTarget_MissingTarget(t *testing.T) {
	fq := &fakeQueryRepo{}
	r := newRouterWithPublic(fq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/lists", nil)
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("want 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestPublicLists_ByTarget_RepoError(t *testing.T) {
	fq := &fakeQueryRepo{errTarget: errors.New("db fail")}
	r := newRouterWithPublic(fq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/lists?target=upbit", nil)
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Fatalf("want 500, got %d body=%s", w.Code, w.Body.String())
	}
}
