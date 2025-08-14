package httpctrl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type slowUpdater struct{
	mu sync.Mutex
	calls int
}
func (u *slowUpdater) Update(ctx context.Context, src, tgt *string) (map[string]int, error) {
	time.Sleep(30 * time.Millisecond)
	u.mu.Lock(); u.calls++; u.mu.Unlock()
	return map[string]int{"okx_to_upbit":1}, nil
}

func TestUpdate_ParallelRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upd := &slowUpdater{}
	r := gin.New()
	r.POST("/update", func(c *gin.Context) {
		res, err := upd.Update(c, nil, nil)
		if err != nil { c.JSON(500, gin.H{"error":err.Error()}); return }
		c.JSON(200, gin.H{"updated":res})
	})

	const N = 10
	var wg sync.WaitGroup
	wg.Add(N)
	for i:=0; i<N; i++ {
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/update", nil)
			r.ServeHTTP(w, req)
			if w.Code != 200 {
				t.Errorf("status %d", w.Code)
			}
		}()
	}
	wg.Wait()
	upd.mu.Lock(); calls := upd.calls; upd.mu.Unlock()
	if calls != N {
		t.Fatalf("concurrent calls=%d want=%d", calls, N)
	}
}
