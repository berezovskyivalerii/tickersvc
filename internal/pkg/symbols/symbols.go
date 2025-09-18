package symbols

import (
	"sort"
	"strings"
)

type Style int

const (
	// BASE + SEP + QUOTE (e.g. "MOG-USDT", "ETH/USDT", "PEPE_USDT")
	StyleBaseSepQuote Style = iota
	// QUOTE + SEP + BASE (Upbit: "BTC-MOG", "KRW-ETH", "USDT-ARB")
	StyleQuoteSepBase
	// CONCAT: BASEQUOTE without sep (Binance: "PEPEUSDT", "ETHBTC")
	StyleConcat
)

var seps = []string{"-", "_", "/"}
var KnownQuotes = []string{
	"USDT","USDC","FDUSD","TUSD","BUSD","USD","EUR","KRW","BTC","ETH","TRY","GBP","JPY",
}

func Split(symbol string, style Style) (string, string, bool) {
	s := strings.TrimSpace(symbol)
	u := strings.ToUpper(s)

	// 1) with sep
	for _, sep := range seps {
		if strings.Contains(u, sep) {
			parts := strings.Split(u, sep)
			if len(parts) != 2 {
				return "", "", false
			}
			switch style {
			case StyleBaseSepQuote:
				return parts[0], parts[1], true
			case StyleQuoteSepBase:
				return parts[1], parts[0], true
			default:
				// unkwnown style — trying to guess
				if looksLikeQuote(parts[1]) {
					return parts[0], parts[1], true
				}
				if looksLikeQuote(parts[0]) {
					return parts[1], parts[0], true
				}
				return parts[0], parts[1], true
			}
		}
	}

	// 2) concat BASEQUOTE — searching for the longest quote suffix
	qs := sortedQuotesByLenDesc()
	for _, q := range qs {
		if strings.HasSuffix(u, q) && len(u) > len(q) {
			base := strings.TrimSuffix(u, q)
			return base, q, true
		}
	}
	return "", "", false
}

func looksLikeQuote(s string) bool {
	for _, q := range KnownQuotes {
		if s == q {
			return true
		}
	}
	return false
}

var quotesSorted []string
func sortedQuotesByLenDesc() []string {
	if quotesSorted == nil {
		quotesSorted = append(quotesSorted, KnownQuotes...)
		sort.Slice(quotesSorted, func(i, j int) bool { return len(quotesSorted[i]) > len(quotesSorted[j]) })
	}
	return quotesSorted
}
