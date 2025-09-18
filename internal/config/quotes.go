package config

import (
	"os"
	"strings"
)

type QuotesConfig struct {
	SourceSpotQuote string
	TargetAllowedQuotes map[string]struct{}
}

func LoadQuotes() QuotesConfig {
	src := getenv("SOURCE_SPOT_QUOTE", "USDT")
	tgt := getenv("TARGET_ALLOWED_QUOTES", "USDT,USD,KRW")
	return QuotesConfig{
		SourceSpotQuote:     strings.ToUpper(strings.TrimSpace(src)),
		TargetAllowedQuotes: toSetCSV(tgt),
	}
}

func toSetCSV(csv string) map[string]struct{} {
	res := make(map[string]struct{}, 8)
	for _, p := range strings.Split(csv, ",") {
		p = strings.ToUpper(strings.TrimSpace(p))
		if p != "" {
			res[p] = struct{}{}
		}
	}
	return res
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
