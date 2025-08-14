package lists

import (
	"reflect"
	"testing"

	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

func spot(base, quote, sym string) dm.Item {
	return dm.Item{Type: dm.TypeSpot, Base: base, Quote: quote, Symbol: sym}
}
func fut(base, quote, sym string) dm.Item {
	return dm.Item{Type: dm.TypeFutures, Base: base, Quote: quote, Symbol: sym}
}

func TestBuildListRows_IgnoreBTCOnly_True(t *testing.T) {
	// Source (OKX/example): AAA, BBB, CCC
	src := []dm.Item{
		spot("AAA", "USDT", "AAA-USDT"),
		fut("AAA", "USDT", "AAA-USDT-SWAP"),
		spot("BBB", "EUR",  "BBB-EUR"),
		spot("CCC", "USDT", "CCC-USDT"),
	}

	// Target (Upbit): AAA to BTC only → DO NOT exclude; BBB to USDT → exclude; CCC absent → do not exclude
	tgt := []dm.Item{
		spot("AAA", "BTC",  "BTC-AAA"),
		spot("BBB", "USDT", "USDT-BBB"),
	}

	got := BuildListRows(src, tgt, "upbit") // ignoreBTCOnly=true (Upbit/Bithumb)
	want := []Row{
		{Spot: "AAA-USDT", Futures: "AAA-USDT-SWAP"},
		{Spot: "CCC-USDT", Futures: "none"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestBuildListRows_IgnoreBTCOnly_False(t *testing.T) {
	// same source
	src := []dm.Item{
		spot("AAA", "USDT", "AAA-USDT"),
		fut("AAA", "USDT", "AAA-USDT-SWAP"),
		spot("CCC", "USDT", "CCC-USDT"),
	}
	// Target (Coinbase): AAA BTC only → now EXCLUDED too; CCC missing → remains
	tgt := []dm.Item{
		spot("AAA", "BTC", "BTC-AAA"),
	}
	got := BuildListRows(src, tgt, "coinbase") // coinbase-behavior
	want := []Row{
		{Spot: "CCC-USDT", Futures: "none"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestFormatAsText(t *testing.T) {
	in := []Row{
		{Spot: "AAA-USDT", Futures: "AAA-USDT-SWAP"},
		{Spot: "CCC-USDT", Futures: "none"},
	}
	txt := FormatAsText(in)
	want := "AAA-USDT, AAA-USDT-SWAP\nCCC-USDT, none\n"
	if txt != want {
		t.Fatalf("got=%q want=%q", txt, want)
	}
}

func TestBuildListRows_PicksBestSpotAndKeepsOrder(t *testing.T) {
	// For AAA is BTC and USDT → take USDT.
	src := []dm.Item{
		spot("AAA", "BTC",  "AAA-BTC"),
		spot("AAA", "USDT", "AAA-USDT"),
	}
	got := BuildListRows(src, nil, "upbit")
	if len(got) != 1 || got[0].Spot != "AAA-USDT" {
		t.Fatalf("bad best-spot pick: %+v", got)
	}
}

func TestBuildListRows_ModeBinance_KeepsIfTargetHasFutures(t *testing.T) {
	src := []dm.Item{
		spot("AAA", "USDT", "AAAUSDT"),
	}
	// On Binance, AAA coin has SPOT and FUTURES - we do not exclude
	tgt := []dm.Item{
		spot("AAA", "USDT", "AAAUSDT"),
		fut("AAA", "USDT", "AAAUSDT-PERP"),
	}
	got := BuildListRows(src, tgt, "binance")
	want := []Row{{Spot: "AAAUSDT", Futures: "none"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}

	// And if Binance only has SPOT, we exclude it
	tgt2 := []dm.Item{ spot("AAA", "USDT", "AAAUSDT") }
	got2 := BuildListRows(src, tgt2, "binance")
	if len(got2) != 0 {
		t.Fatalf("expected empty, got=%v", got2)
	}
}
