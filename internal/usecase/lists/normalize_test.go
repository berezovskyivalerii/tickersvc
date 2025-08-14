package lists_test

import (
	"reflect"
	"testing"

	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
	"github.com/berezovskyivalerii/tickersvc/internal/usecase/lists"
)

func spot(base, quote, symbol string) dm.Item {
	return dm.Item{
		ExchangeID: 1, Type: dm.TypeSpot,
		Base: base, Quote: quote, Symbol: symbol, Active: true,
	}
}
func fut(base, quote, symbol string) dm.Item {
	return dm.Item{
		ExchangeID: 1, Type: dm.TypeFutures,
		Base: base, Quote: quote, Symbol: symbol, Active: true,
	}
}

func TestBuildPresence_BTCOnlyVsNonBTC(t *testing.T) {
	input := []dm.Item{
		spot("AAA", "BTC", "BTC-AAA"),       // only BTC
		spot("BBB", "USDT", "BBB-USDT"),     // non-BTC
		spot("CCC", "KRW", "CCC-KRW"),       // non-BTC
		fut("AAA", "USDT", "AAAUSDT-PERP"),  // should be ignored here
	}
	p := lists.BuildPresence(input)

	if got := p["AAA"]; !(got.HasAny && got.HasBTC && !got.HasNonBTC) {
		t.Fatalf("AAA presence wrong: %+v", got)
	}
	if got := p["BBB"]; !(got.HasAny && !got.HasBTC && got.HasNonBTC) {
		t.Fatalf("BBB presence wrong: %+v", got)
	}
	if got := p["CCC"]; !(got.HasAny && !got.HasBTC && got.HasNonBTC) {
		t.Fatalf("CCC presence wrong: %+v", got)
	}

	// quotes set recorded (case-insensitive → stored upper)
	if _, ok := p["BBB"].Quotes["USDT"]; !ok {
		t.Fatalf("BBB USDT not recorded in quotes")
	}
	if _, ok := p["AAA"].Quotes["BTC"]; !ok {
		t.Fatalf("AAA BTC not recorded in quotes")
	}
}

func TestBuildSourceIndex_PickBestSpotAndFutures(t *testing.T) {
	input := []dm.Item{
		// AAA has BTC and USDT spot → pick USDT by priority, has futures too
		spot("AAA", "BTC", "AAABTC"),
		spot("AAA", "USDT", "AAAUSDT"),
		fut("AAA", "USDT", "AAAUSDT-PERP"),

		// BBB only EUR → pick EUR
		spot("BBB", "EUR", "BBBEUR"),

		// cCc only BTC (lowercase to test normalization) → pick BTC, no futures
		spot("cCc", "btc", "BTC-CCC"),

		// DDD only futures (no spot) → should be absent from index
		fut("DDD", "USDT", "DDDUSDT-PERP"),
	}

	idx := lists.BuildSourceIndex(input)

	if si, ok := idx["AAA"]; !ok || si.SpotSymbol != "AAAUSDT" {
		t.Fatalf("AAA spot chosen wrong: %+v", si)
	}
	if si := idx["AAA"]; si.FuturesSymbol == nil || *si.FuturesSymbol != "AAAUSDT-PERP" {
		t.Fatalf("AAA futures missing/wrong: %+v", si)
	}

	if si, ok := idx["BBB"]; !ok || si.SpotSymbol != "BBBEUR" {
		t.Fatalf("BBB spot chosen wrong: %+v", si)
	}

	// key should be normalized to upper base
	if si, ok := idx["CCC"]; !ok || si.SpotSymbol != "BTC-CCC" {
		t.Fatalf("CCC spot chosen wrong or missing: %+v", si)
	}
	if si := idx["CCC"]; si.FuturesSymbol != nil {
		t.Fatalf("CCC should not have futures: %+v", si)
	}

	if _, ok := idx["DDD"]; ok {
		t.Fatalf("DDD must not be present without spot")
	}
}

func TestMakeList_ModesAndSorting(t *testing.T) {
	// source index (already normalized)
	srcIdx := map[string]lists.SourceInfo{
		"AAA": {SpotSymbol: "AAAUSDT", FuturesSymbol: strPtr("AAAUSDT-PERP")},
		"BBB": {SpotSymbol: "BBBEUR", FuturesSymbol: nil},
		"CCC": {SpotSymbol: "CCCUSDT", FuturesSymbol: nil},
	}

	// target presence variants
	// upbit/bithumb: AAA only BTC → keep; BBB has USDT → exclude; CCC absent → keep
	upbitPresence := map[string]lists.Presence{
		"AAA": {HasAny: true, HasBTC: true, HasNonBTC: false, Quotes: map[string]struct{}{"BTC": {}}},
		"BBB": {HasAny: true, HasBTC: false, HasNonBTC: true, Quotes: map[string]struct{}{"USDT": {}}},
		// CCC missing
	}

	// coinbase: any presence excludes
	coinPresence := map[string]lists.Presence{
		"AAA": {HasAny: true, HasBTC: true, HasNonBTC: false, Quotes: map[string]struct{}{"BTC": {}}},
		"BBB": {HasAny: true, HasBTC: false, HasNonBTC: true, Quotes: map[string]struct{}{"EUR": {}}},
		// CCC missing
	}

	// upbit mode
	rows := lists.MakeList(srcIdx, upbitPresence, "upbit")
	wantUpbit := []lists.Row{
		{Spot: "AAAUSDT", Futures: "AAAUSDT-PERP"},
		{Spot: "CCCUSDT", Futures: "none"},
	}
	if !reflect.DeepEqual(rows, wantUpbit) {
		t.Fatalf("upbit rows:\n got=%v\nwant=%v", rows, wantUpbit)
	}

	// bithumb mode behaves the same rule as upbit
	rows = lists.MakeList(srcIdx, upbitPresence, "bithumb")
	if !reflect.DeepEqual(rows, wantUpbit) {
		t.Fatalf("bithumb rows:\n got=%v\nwant=%v", rows, wantUpbit)
	}

	// coinbase mode: exclude any presence → only CCC remains
	rows = lists.MakeList(srcIdx, coinPresence, "coinbase")
	wantCB := []lists.Row{
		{Spot: "CCCUSDT", Futures: "none"},
	}
	if !reflect.DeepEqual(rows, wantCB) {
		t.Fatalf("coinbase rows:\n got=%v\nwant=%v", rows, wantCB)
	}

	// default mode == coinbase
	rows = lists.MakeList(srcIdx, coinPresence, "")
	if !reflect.DeepEqual(rows, wantCB) {
		t.Fatalf("default rows:\n got=%v\nwant=%v", rows, wantCB)
	}

	// verify sorting by Spot lexicographically
	// (AAA... comes before CCC..., already checked by upbit want order)
}

func strPtr(s string) *string { return &s }
