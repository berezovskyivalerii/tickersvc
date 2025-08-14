package app

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"

	httpctrl "github.com/berezovskyivalerii/tickersvc/internal/adapter/controller/http"
	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/dbping"
	"github.com/berezovskyivalerii/tickersvc/internal/config"
	domain "github.com/berezovskyivalerii/tickersvc/internal/domain/health"
	httpinfra "github.com/berezovskyivalerii/tickersvc/internal/infra/http"
	"github.com/berezovskyivalerii/tickersvc/internal/infra/store"
	usehealth "github.com/berezovskyivalerii/tickersvc/internal/usecase/health"
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

	var pingers []domain.Pinger
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

	return router, nil
}
