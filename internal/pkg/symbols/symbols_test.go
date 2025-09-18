package symbols_test

import (
	"testing"

	"github.com/berezovskyivalerii/tickersvc/internal/pkg/symbols"
)

func TestSplit_BaseSepQuote(t *testing.T) {
	base, quote, ok := symbols.Split("MOG-USDT", symbols.StyleBaseSepQuote)
	if !ok || base != "MOG" || quote != "USDT" {
		t.Fatalf("got %q %q ok=%v", base, quote, ok)
	}
	base, quote, ok = symbols.Split("PEPE_USDC", symbols.StyleBaseSepQuote)
	if !ok || base != "PEPE" || quote != "USDC" { t.Fatal() }
	base, quote, ok = symbols.Split("ETH/BTC", symbols.StyleBaseSepQuote)
	if !ok || base != "ETH" || quote != "BTC" { t.Fatal() }
}

func TestSplit_QuoteSepBase_Upbit(t *testing.T) {
	base, quote, ok := symbols.Split("BTC-MOG", symbols.StyleQuoteSepBase)
	if !ok || base != "MOG" || quote != "BTC" { t.Fatal() }
	base, quote, ok = symbols.Split("KRW-ETH", symbols.StyleQuoteSepBase)
	if !ok || base != "ETH" || quote != "KRW" { t.Fatal() }
}

func TestSplit_Concat_Binance(t *testing.T) {
	base, quote, ok := symbols.Split("PEPEUSDT", symbols.StyleConcat)
	if !ok || base != "PEPE" || quote != "USDT" { t.Fatal() }
	base, quote, ok = symbols.Split("ETHBTC", symbols.StyleConcat)
	if !ok || base != "ETH" || quote != "BTC" { t.Fatal() }
}

func TestSplit_Heuristics_NoStyle(t *testing.T) {
	base, quote, ok := symbols.Split("ARB-USDT", 99)
	if !ok || base != "ARB" || quote != "USDT" { t.Fatal() }

	base, quote, ok = symbols.Split("ETHBTC", 99)
	if !ok || base != "ETH" || quote != "BTC" { t.Fatal() }
}

func TestSplit_Edge_PrefixMultiplier(t *testing.T) {
	base, quote, ok := symbols.Split("1000PEPEUSDT", symbols.StyleConcat)
	if !ok || base != "1000PEPE" || quote != "USDT" { t.Fatalf("got %s/%s", base, quote) }
}
