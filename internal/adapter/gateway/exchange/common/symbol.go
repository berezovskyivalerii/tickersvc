package common

import "strings"

// TrimPerpSuffix возвращает core-символ без окончаний PERP/SWAP и признак фьючерса.
func TrimPerpSuffix(sym string) (core string, perp bool) {
	s := sym
	for _, suf := range []string{
		"-PERP", "PERP", "_PERP",
		"-SWAP", "SWAP",
		"-USDT-PERP", "-USD-PERP",
	} {
		if strings.HasSuffix(s, suf) {
			return strings.TrimSuffix(s, suf), true
		}
	}
	return s, false
}

// SplitKnownQuote делит склеенный символ по известным котировочным валютам.
// Не «исправляет» числа в начале (1000PEPE остаётся как есть).
func SplitKnownQuote(sym string, quotes []string) (base, quote string, ok bool) {
	s := sym
	if core, _ := TrimPerpSuffix(s); core != "" {
		s = core
	}
	for _, q := range quotes {
		if strings.HasSuffix(s, q) {
			base = s[:len(s)-len(q)]
			quote = q
			if base != "" {
				return base, quote, true
			}
		}
	}
	return "", "", false
}

// CommonQuoteSet — разумный набор для сплита binance-стиля.
var CommonQuoteSet = []string{
	"USDT", "USDC", "USD", "FDUSD", "BUSD",
	"BTC", "ETH", "BNB", "EUR", "TRY", "KRW", "BRL", "TUSD",
}
