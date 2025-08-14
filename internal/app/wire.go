package app

import (
	"log/slog"
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
	"github.com/berezovskyivalerii/tickersvc/internal/infra/logx"
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

	// --- /health ---
	var pingers []healthdom.Pinger
	pingers = append(pingers, dbping.DBPing{DB: db})
	uc := &usehealth.ReadinessInteractor{
		Pingers:   pingers,
		Version:   config.Version,
		Commit:    config.Commit,
		BuildTime: config.BuildTime,
		StartedAt: config.StartedAt,
		Clock:     usehealth.SysClock{},
		Timeout:   500 * time.Millisecond,
	}

	router := httpinfra.NewRouter()
	health := httpctrl.NewHealthController(httpctrl.ReadinessRunner{UC: uc})
	health.Register(router)

	// --- Repos / Use-cases ---
	// markets
	marketsRepo := pgrepo.NewMarketsRepo(db)
	exclude := map[string]bool{}
	for _, s := range strings.Split(os.Getenv("EXCLUDE_EXCHANGES"), ",") {
		s = strings.TrimSpace(strings.ToLower(s))
		if s != "" { exclude[s] = true }
	}
	
	all := map[string]marketsdom.Fetcher{
		"binance":   exbinance.New(),
		"bybit":     exbybit.New(),
		"okx":       exokx.New(),
		"upbit":     exupbit.New(),
		"coinbase":  excoin.New(),
		"bithumb":   exbithumb.New(),
		"robinhood": exrobin.New(),
	}
	var fetchers []marketsdom.Fetcher
	for slug, f := range all {
		if exclude[slug] { continue }
		fetchers = append(fetchers, f)
	}
	
	logger := logx.New()
	slog.SetDefault(logger)
	
	// Orchestrator sync
	marketsOrc := &marketsuc.Orchestrator{
		Repo:     marketsRepo,
		Fetchers: fetchers,
		Timeout:  45 * time.Second,
		Logger:   logger,
	}

	// lists: defs + saver + reader
	defsRepo := pgrepo.NewListDefsRepo(db)
	listsSaver := pgrepo.NewListsRepo(db)
	listsReader := pgrepo.NewListsQueryRepo(db)

	updater := &listsuc.Updater{
		Defs:  defsRepo,
		MRepo: marketsRepo,
		Saver: listsSaver,
	}

	// --- Admin: manual sync (diagnosis) ---
	admin := router.Group("/admin")
	admin.POST("/markets/sync", func(c *gin.Context) {
		summary, err := marketsOrc.RunAll(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"summary": summary})
	})

	// --- /update: sync -> lists rebuild ---
	// ?source=okx&target=upbit  (both params optional)
	router.POST("/update", func(c *gin.Context) {
		// 1) всегда подтягиваем свежие рынки
		summary, err := marketsOrc.RunAll(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// 2) lists rebuild
		var srcPtr, tgtPtr *string
		if q := strings.TrimSpace(c.Query("source")); q != "" {
			s := strings.Split(q, ",")[0]
			srcPtr = &s
		}
		if q := strings.TrimSpace(c.Query("target")); q != "" {
			t := strings.Split(q, ",")[0]
			tgtPtr = &t
		}

		updated, err := updater.Update(c.Request.Context(), srcPtr, tgtPtr)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error(), "markets_sync": summary})
			return
		}
		c.JSON(200, gin.H{
			"markets_sync":  summary, // map[exchangeID][added,updated,archived]
			"lists_updated": updated, // map[slug]inserted_count
		})
	})

	listsCtrl := httpctrl.NewListsController(listsReader)
	listsCtrl.Register(router)

	return router, nil
}
