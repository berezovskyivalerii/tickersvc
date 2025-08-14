package httpctrl

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	ldom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
)

type fakeQueryRepo struct {
	all      map[string]map[string][]string
	target   map[string]map[string][]string // target -> source -> lines
	slugData map[string][]string            // slug -> lines
	errAll   error
	errTgt   error
	errSlug  error
}

func (f *fakeQueryRepo) GetAllText(ctx context.Context) (map[string]map[string][]string, error) {
	return f.all, f.errAll
}
func (f *fakeQueryRepo) GetTextByTarget(ctx context.Context, targetSlug string) (map[string][]string, error) {
	return f.target[targetSlug], f.errTgt
}
func (f *fakeQueryRepo) GetTextBySlug(ctx context.Context, slug string) ([]string, error) {
	return f.slugData[slug], f.errSlug
}

var _ ldom.QueryRepo = (*fakeQueryRepo)(nil)

func TestLists_All_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fq := &fakeQueryRepo{
		all: map[string]map[string][]string{
			"upbit":   {"okx": {"AAA-USDT, AAA-USDT-SWAP", "CCC-USDT, none"}},
			"bithumb": {"okx": {"BBB-USDT, none"}},
		},
	}
	r := gin.New()
	NewListsController(fq).Register(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/lists", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var got map[string]map[string][]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(got["upbit"]["okx"]) != 2 {
		t.Fatalf("bad payload: %#v", got)
	}
}

func TestLists_All_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fq := &fakeQueryRepo{errAll: errors.New("db down")}
	r := gin.New()
	NewListsController(fq).Register(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/lists", nil)
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

func TestLists_ByTarget_OK_EmptyAndNonEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fq := &fakeQueryRepo{
		target: map[string]map[string][]string{
			"upbit":   {"okx": {"AAA-USDT, AAA-USDT-SWAP"}},
			"bithumb": {}, // пусто — допустимо
		},
	}
	r := gin.New()
	NewListsController(fq).Register(r)

	// upbit — есть данные
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/lists/upbit", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var got map[string][]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(got["okx"]) != 1 {
		t.Fatalf("bad payload: %#v", got)
	}

	// bithumb — пустой объект
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/lists/bithumb", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != 200 {
		t.Fatalf("status2=%d body=%s", w2.Code, w2.Body.String())
	}
	var got2 map[string][]string
	if err := json.Unmarshal(w2.Body.Bytes(), &got2); err != nil {
		t.Fatalf("json2: %v", err)
	}
	if len(got2) != 0 {
		t.Fatalf("expected empty map, got %#v", got2)
	}
}

func TestLists_ByTarget_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fq := &fakeQueryRepo{errTgt: errors.New("boom"), target: map[string]map[string][]string{}}
	r := gin.New()
	NewListsController(fq).Register(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/lists/upbit", nil)
	r.ServeHTTP(w, req)
	if w.Code != 500 {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

func TestList_Single_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fq := &fakeQueryRepo{
		slugData: map[string][]string{
			"okx_to_upbit": {"AAA-USDT, AAA-USDT-SWAP", "CCC-USDT, none"},
		},
	}
	r := gin.New()
	NewListsController(fq).Register(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/list?source=okx&target=upbit", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct == "" || ct[:10] != "text/plain" {
		t.Fatalf("content-type=%s", ct)
	}
	if got := w.Body.String(); got != "AAA-USDT, AAA-USDT-SWAP\nCCC-USDT, none\n" {
		t.Fatalf("body:\n%s", got)
	}
}

func TestList_Single_MissingParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fq := &fakeQueryRepo{slugData: map[string][]string{}}
	r := gin.New()
	NewListsController(fq).Register(r)

	cases := []string{
		"/list?source=okx",         // нет target
		"/list?target=upbit",       // нет source
		"/list",                    // нет обоих
		"/list?source=&target=",    // пустые
	}
	for _, path := range cases {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("%s: want 400, got %d body=%s", path, w.Code, w.Body.String())
		}
	}
}

func TestList_Single_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fq := &fakeQueryRepo{errSlug: errors.New("db fail"), slugData: map[string][]string{}}
	r := gin.New()
	NewListsController(fq).Register(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/list?source=okx&target=upbit", nil)
	r.ServeHTTP(w, req)
	if w.Code != 500 {
		t.Fatalf("want 500, got %d", w.Code)
	}
}
