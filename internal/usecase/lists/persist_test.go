package lists

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"
	"time"

	pg "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/postgres"
	//listsdom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
	"github.com/berezovskyivalerii/tickersvc/internal/infra/store"
)

func TestRowsToItems(t *testing.T) {
	in := []Row{
		{Spot: "AAA-USDT", Futures: "AAA-USDT-SWAP"},
		{Spot: "BBB-USDT", Futures: "none"},
		{Spot: "CCC-USDT", Futures: ""},
	}
	got := RowsToItems(in)

	if len(got) != 3 {
		t.Fatalf("len=%d", len(got))
	}
	// 1
	if got[0].Spot != "AAA-USDT" || got[0].Futures == nil || *got[0].Futures != "AAA-USDT-SWAP" {
		t.Fatalf("row0 bad: %+v", got[0])
	}
	// 2
	if got[1].Spot != "BBB-USDT" || got[1].Futures != nil {
		t.Fatalf("row1 bad: %+v", got[1])
	}
	// 3
	if got[2].Spot != "CCC-USDT" || got[2].Futures != nil {
		t.Fatalf("row2 bad: %+v", got[2])
	}
}

func TestRebuildAndSave_Integration(t *testing.T) {
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

	// исходные данные (источник: OKX, цель: Upbit)
	source := []dm.Item{
		spot("AAA", "USDT", "AAA-USDT"),
		fut("AAA", "USDT", "AAA-USDT-SWAP"),
		spot("BBB", "EUR", "BBB-EUR"),
		spot("CCC", "USDT", "CCC-USDT"),
	}
	target := []dm.Item{
		spot("AAA", "BTC", "BTC-AAA"),   // BTC-only → не исключаем в режиме upbit
		spot("BBB", "USDT", "USDT-BBB"), // non-BTC → исключаем BBB
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// первый пересчёт и сохранение
	inserted, err := RebuildAndSave(ctx, repo, "okx_to_upbit", source, target, "upbit")
	if err != nil {
		t.Fatalf("RebuildAndSave err: %v", err)
	}
	if inserted != 2 {
		t.Fatalf("inserted=%d want=2", inserted)
	}

	type row struct{ Spot string; Futures *string }
	rows := mustSelectBySlug(t, db, "okx_to_upbit")
	want := []row{
		{Spot: "AAA-USDT", Futures: strPtr("AAA-USDT-SWAP")},
		{Spot: "CCC-USDT", Futures: nil},
	}
	gotVals := make([][2]string, 0, len(rows))
	for _, r := range rows {
		fut := ""
		if r.Futures != nil { fut = *r.Futures }
		gotVals = append(gotVals, [2]string{r.Spot, fut})
	}
	wantVals := make([][2]string, 0, len(want))
	for _, r := range want {
		fut := ""
		if r.Futures != nil { fut = *r.Futures }
		wantVals = append(wantVals, [2]string{r.Spot, fut})
	}
	if !reflect.DeepEqual(gotVals, wantVals) {
		t.Fatalf("rows got=%v want=%v", gotVals, wantVals)
	}

	// updated_at должен обновиться на втором вызове
	var before time.Time
	if err := db.QueryRowContext(ctx, `SELECT updated_at FROM list_defs WHERE slug=$1`, "okx_to_upbit").Scan(&before); err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Millisecond) // небольшой интервал, чтобы метка времени изменилась

	// второй пересчёт с другим набором (замена содержимого)
	source2 := []dm.Item{
		spot("ZZZ", "USDT", "ZZZ-USDT"),
		fut("ZZZ", "USDT", "ZZZ-USDT-SWAP"),
	}
	target2 := []dm.Item{} // на целевой нет
	inserted2, err := RebuildAndSave(ctx, repo, "okx_to_upbit", source2, target2, "upbit")
	if err != nil {
		t.Fatalf("RebuildAndSave#2 err: %v", err)
	}
	if inserted2 != 1 {
		t.Fatalf("inserted2=%d want=1", inserted2)
	}
	_ = mustSelectBySlug(t, db, "okx_to_upbit")
	_ = []row{
		{Spot: "ZZZ-USDT", Futures: strPtr("ZZZ-USDT-SWAP")},
	}
	gotVals = make([][2]string, 0, len(rows))
	for _, r := range rows {
		fut := ""
		if r.Futures != nil { fut = *r.Futures }
		gotVals = append(gotVals, [2]string{r.Spot, fut})
	}
	wantVals = make([][2]string, 0, len(want))
	for _, r := range want {
		fut := ""
		if r.Futures != nil { fut = *r.Futures }
		wantVals = append(wantVals, [2]string{r.Spot, fut})
	}
	if !reflect.DeepEqual(gotVals, wantVals) {
		t.Fatalf("rows got=%v want=%v", gotVals, wantVals)
	}

	var after time.Time
	if err := db.QueryRowContext(ctx, `SELECT updated_at FROM list_defs WHERE slug=$1`, "okx_to_upbit").Scan(&after); err != nil {
		t.Fatal(err)
	}
	if !after.After(before) {
		t.Fatalf("updated_at not advanced: before=%v after=%v", before, after)
	}
}

func mustSelectBySlug(t *testing.T, db *sql.DB, slug string) []struct {
	Spot string
	Futures *string
} {
	t.Helper()
	q := `
SELECT li.spot_symbol, li.futures_symbol
FROM list_items li
JOIN list_defs ld ON ld.id = li.list_id
WHERE ld.slug = $1
ORDER BY li.spot_symbol`
	rows, err := db.Query(q, slug)
	if err != nil {
		t.Fatalf("select list_items: %v", err)
	}
	defer rows.Close()
	var out []struct {
		Spot string
		Futures *string
	}
	for rows.Next() {
		var spot string
		var fut *string
		if err := rows.Scan(&spot, &fut); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, struct {
			Spot    string
			Futures *string
		}{Spot: spot, Futures: fut})
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}
	return out
}

func strPtr(s string) *string { return &s }
