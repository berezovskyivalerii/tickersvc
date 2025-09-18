package main

import (
	"log"
	"os"

	"github.com/berezovskyivalerii/tickersvc/internal/app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	r, err := app.Build()
	if err != nil { log.Fatal(err) }

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
