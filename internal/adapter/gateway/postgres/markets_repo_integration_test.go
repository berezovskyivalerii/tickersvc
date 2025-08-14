package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/postgres"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
	"github.com/berezovskyivalerii/tickersvc/internal/infra/store"
)

func TestMarketsRepo_SyncSnapshot(t *testing.T) {
    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        t.Skip("DB_DSN not set; integration test skipped")
    }
    db, err := store.OpenPostgres(dsn)
    if err != nil { t.Fatal(err) }
    defer db.Close()

    repo := postgres.NewMarketsRepo(db)

    // уникальный суффикс, чтобы точно были вставки
    suf := time.Now().UnixNano()
    exID := int16(3) // okx есть в сид-миграции

    items := []dm.Item{
        {ExchangeID: exID, Type: dm.TypeSpot,    Symbol: fmt.Sprintf("T%v-USDT", suf), Base: fmt.Sprintf("T%v", suf), Quote: "USDT", Active: true},
        {ExchangeID: exID, Type: dm.TypeFutures, Symbol: fmt.Sprintf("T%v-USDT-SWAP", suf), Base: fmt.Sprintf("T%v", suf), Quote: "USDT", Active: true},
    }

    added, updated, _, err := repo.SyncSnapshot(context.Background(), exID, items)
    if err != nil { t.Fatal(err) }
    if added == 0 && updated == 0 {
        t.Fatalf("expected insert/update > 0, got a=%d u=%d", added, updated)
    }

    // второй прогон — должны быть 0 добавлено, возможно 0 обновлено, 0 архивов
    added2, updated2, archived2, err := repo.SyncSnapshot(context.Background(), exID, items)
    if err != nil { t.Fatal(err) }
    if added2 != 0 || archived2 != 0 {
        t.Fatalf("idempotency fail: a2=%d u2=%d d2=%d", added2, updated2, archived2)
    }
}

