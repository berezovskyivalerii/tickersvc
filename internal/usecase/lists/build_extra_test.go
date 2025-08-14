package lists

import (
	"reflect"
	"testing"

	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

func s(base, quote, sym string) dm.Item { return dm.Item{Type: dm.TypeSpot, Base: base, Quote: quote, Symbol: sym, Active: true} }
func f(base, quote, sym string) dm.Item { return dm.Item{Type: dm.TypeFutures, Base: base, Quote: quote, Symbol: sym, Active: true} }

func TestBuildListRows_OnlyFuturesOnSource_Ignored(t *testing.T) {
	src := []dm.Item{ f("AAA","USDT","AAAUSDT-PERP") } // нет спота → монета не попадёт в список
	tgt := []dm.Item{} // целевая пустая
	got := BuildListRows(src, tgt, "coinbase")
	if len(got) != 0 {
		t.Fatalf("must be empty, got=%v", got)
	}
}

func TestBuildListRows_TargetHasOnlyFutures_CoinbaseMode_Excludes(t *testing.T) {
	src := []dm.Item{ s("AAA","USDT","AAAUSDT") }
	// На целевой есть только фьючерсы по AAA — в coinbase-режиме мы исключаем по "любому присутствию"? Нет: у нас логика считает только spot для coinbase.
	// Спец-проверка: HasAnySpot=false → не исключаем.
	tgt := []dm.Item{ f("AAA","USDT","AAAUSDT-PERP") }
	got := BuildListRows(src, tgt, "coinbase")
	want := []Row{{Spot:"AAAUSDT", Futures:"none"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestBuildListRows_ModeBinance_KeepIfTargetHasFutures(t *testing.T) {
	src := []dm.Item{ s("AAA","USDT","AAAUSDT") }
	tgt := []dm.Item{ s("AAA","USDT","AAAUSDT"), f("AAA","USDT","AAAUSDT-PERP") }
	got := BuildListRows(src, tgt, "binance")
	want := []Row{{Spot:"AAAUSDT", Futures:"none"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
	tgt2 := []dm.Item{ s("AAA","USDT","AAAUSDT") } // только спот → исключаем
	got2 := BuildListRows(src, tgt2, "binance")
	if len(got2) != 0 { t.Fatalf("expected empty, got=%v", got2) }
}

func TestBuildListRows_Upbit_BTCOnlyIsKept_CaseInsensitive(t *testing.T) {
	src := []dm.Item{ s("aaa","usdt","AAA-USDT") }
	tgt := []dm.Item{ s("AAA","btc","BTC-AAA") } // только BTC на целевой
	got := BuildListRows(src, tgt, "upbit")
	want := []Row{{Spot:"AAA-USDT", Futures:"none"}}
	if !reflect.DeepEqual(got, want) { t.Fatalf("got=%v want=%v", got, want) }
}
