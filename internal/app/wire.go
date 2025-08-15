package app

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	httpctrl "github.com/berezovskyivalerii/tickersvc/internal/adapter/controller/http"
	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/dbping"
	exbinance "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/binance"
	exbithumb "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/bithumb"
	exbybit "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/bybit"
	excoin "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/coinbase"
	exokx "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/okx"
	exrobin "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/robinhood"
	exupbit "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/upbit"
	pgrepo "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/postgres"
	"github.com/berezovskyivalerii/tickersvc/internal/config"
	healthdom "github.com/berezovskyivalerii/tickersvc/internal/domain/health"
	marketsdom "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
	httpinfra "github.com/berezovskyivalerii/tickersvc/internal/infra/http"
	adminauth "github.com/berezovskyivalerii/tickersvc/internal/infra/http/mw/adminauth"
	"github.com/berezovskyivalerii/tickersvc/internal/infra/store"
	usehealth "github.com/berezovskyivalerii/tickersvc/internal/usecase/health"
	listsuc "github.com/berezovskyivalerii/tickersvc/internal/usecase/lists"
	marketsuc "github.com/berezovskyivalerii/tickersvc/internal/usecase/markets"
)

type envErr string

func (e envErr) Error() string { return "missing env: " + string(e) }
func ErrEnv(name string) error { return envErr(name) }

func Build() (*gin.Engine, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, ErrEnv("DB_DSN")
	}

	db, err := store.OpenPostgres(dsn)
	if err != nil {
		return nil, err
	}

	// --- Health (/health) ---
	var pingers []healthdom.Pinger
	pingers = append(pingers, dbping.DBPing{DB: db})

	ucHealth := &usehealth.ReadinessInteractor{
		Pingers:   pingers,
		Version:   config.Version,
		Commit:    config.Commit,
		BuildTime: config.BuildTime,
		StartedAt: config.StartedAt,
		Clock:     usehealth.SysClock{},
		Timeout:   500 * time.Millisecond,
	}

	router := httpinfra.NewRouter()
	_ = router.SetTrustedProxies(nil)

	health := httpctrl.NewHealthController(httpctrl.ReadinessRunner{UC: ucHealth})
	health.Register(router)

	// --- Repos ---
	marketsRepo := pgrepo.NewMarketsRepo(db)
	defsRepo := pgrepo.NewListDefsRepo(db)
	listsSaver := pgrepo.NewListsRepo(db)
	listsReader := pgrepo.NewListsQueryRepo(db)
	exchangesRepo := pgrepo.NewExchangesRepo(db)

	// --- Active exchanges from DB + ENV-excludes ---
	actMap, err := exchangesRepo.ActiveMap(context.Background())
	if err != nil {
		return nil, err
	}
	exclude := map[string]bool{}
	if v := os.Getenv("EXCLUDE_EXCHANGES"); v != "" {
		for _, s := range strings.Split(v, ",") {
			s = strings.TrimSpace(strings.ToLower(s))
			if s != "" {
				exclude[s] = true
			}
		}
	}

	// All fetchers
	allFetchers := []marketsdom.Fetcher{
		exbinance.New(),
		exbybit.New(),
		exokx.New(),
		exupbit.New(),
		excoin.New(),
		exbithumb.New(),
		exrobin.New(),
	}

	// Apply activities filters and exludes
	var fetchers []marketsdom.Fetcher
	for _, f := range allFetchers {
		if exclude[f.Name()] {
			continue
		}
		if on, ok := actMap[f.ExchangeID()]; ok && !on {
			continue
		}
		fetchers = append(fetchers, f)
	}

	marketsOrc := &marketsuc.Orchestrator{
		Repo:     marketsRepo,
		Fetchers: fetchers,
		Timeout:  45 * time.Second,
	}

	listsInteractor := &listsuc.Interactor{
		Defs:    defsRepo,
		Markets: marketsRepo,
		Lists:   listsSaver,
	}

	// Public lists
	pub := httpctrl.NewPublicListsController(listsReader)
	pub.Register(router) // /api/lists/:slug и /api/lists?target=... [&as_text=1]

	router.POST("/update", func(c *gin.Context) {
		summary, err := marketsOrc.RunAll(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		var srcPtr, tgtPtr *string
		if q := strings.TrimSpace(c.Query("source")); q != "" {
			s := strings.TrimSpace(strings.Split(q, ",")[0])
			if s != "" {
				srcPtr = &s
			}
		}
		if q := strings.TrimSpace(c.Query("target")); q != "" {
			t := strings.TrimSpace(strings.Split(q, ",")[0])
			if t != "" {
				tgtPtr = &t
			}
		}

		updated, err := listsInteractor.BuildAndSaveFiltered(c.Request.Context(), srcPtr, tgtPtr)
		if err != nil {
			c.JSON(500, gin.H{
				"error":        err.Error(),
				"markets_sync": summary,
			})
			return
		}
		c.JSON(200, gin.H{
			"markets_sync":  summary,
			"lists_updated": updated,
		})
	})

	admin := router.Group("/admin", adminauth.NewFromEnv().Handler())
	admin.POST("/markets/sync", func(c *gin.Context) {
		summary, err := marketsOrc.RunAll(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"summary": summary})
	})

	return router, nil
}
