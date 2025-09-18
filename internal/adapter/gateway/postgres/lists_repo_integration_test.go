package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"
	"time"

	pg "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/postgres"
	listsdom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
	"github.com/berezovskyivalerii/tickersvc/internal/infra/store"
)

func TestListsRepo_ReplaceBySlug_Basic(t *testing.T) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("DB_DSN not set; integration test skipped")
	}
	db, err := store.OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := pg.NewListsRepo(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1) inserting two elements
	items := []listsdom.Item{
		{Spot: "AAA-USDT", Futures: strPtr("AAA-USDT-SWAP")},
		{Spot: "BBB-USDT", Futures: nil},
	}
	ins, err := repo.ReplaceBySlug(ctx, "okx_to_bithumb", items)
	if err != nil {
		t.Fatalf("ReplaceBySlug err: %v", err)
	}
	if ins != 2 {
		t.Fatalf("inserted=%d want=2", ins)
	}
	got := selectBySlug(t, db, "okx_to_bithumb")
	want := []row{
		{"AAA-USDT", strPtr("AAA-USDT-SWAP")},
		{"BBB-USDT", nil},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}

	// 2) replacement (truncate-insert)
	items2 := []listsdom.Item{
		{Spot: "ZZZ-USDT", Futures: strPtr("ZZZ-USDT-SWAP")},
	}
	ins2, err := repo.ReplaceBySlug(ctx, "okx_to_bithumb", items2)
	if err != nil {
		t.Fatalf("ReplaceBySlug#2 err: %v", err)
	}
	if ins2 != 1 {
		t.Fatalf("inserted2=%d want=1", ins2)
	}
	got2 := selectBySlug(t, db, "okx_to_bithumb")
	want2 := []row{{"ZZZ-USDT", strPtr("ZZZ-USDT-SWAP")}}
	if !reflect.DeepEqual(got2, want2) {
		t.Fatalf("got2=%v want2=%v", got2, want2)
	}
}

func TestListsRepo_ReplaceBySlug_NotFound(t *testing.T) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("DB_DSN not set; integration test skipped")
	}
	db, err := store.OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := pg.NewListsRepo(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = repo.ReplaceBySlug(ctx, "no_such_slug", nil)
	if err == nil {
		t.Fatal("expected error for unknown slug")
	}
}

type row struct {
	Spot string
	Fut  *string
}

func selectBySlug(t *testing.T, db *sql.DB, slug string) []row {
	t.Helper()
	const q = `
SELECT li.spot_symbol, li.futures_symbol
FROM list_items li
JOIN list_defs ld ON ld.id = li.list_id
WHERE ld.slug = $1
ORDER BY li.spot_symbol`
	rs, err := db.Query(q, slug)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	defer rs.Close()
	var out []row
	for rs.Next() {
		var s string
		var f *string
		if err := rs.Scan(&s, &f); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, row{Spot: s, Fut: f})
	}
	if err := rs.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}
	return out
}

func strPtr(s string) *string { return &s }
