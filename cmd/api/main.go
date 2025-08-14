package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	httpapi "github.com/berezovskyivalerii/tickersvc/internal/http"
	"github.com/berezovskyivalerii/tickersvc/internal/config"
	"github.com/berezovskyivalerii/tickersvc/internal/store"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is required")
	}

	db, err := store.OpenPostgres(dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	build := config.NewBuildInfo()

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	health := httpapi.NewHealthHandler(db, build)
	health.Register(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
