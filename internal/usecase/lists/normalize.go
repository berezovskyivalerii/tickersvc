package lists

import (
	"sort"
	"strings"

	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

type Presence struct {
	Quotes    map[string]struct{} // what is quoted at the target
	HasAny    bool
	HasBTC    bool
	HasNonBTC bool
}

type SourceInfo struct {
	SpotSymbol    string  // chosen source's spot ticker
	FuturesSymbol *string // future's ticker or nil
}

type Row struct {
	Spot    string
	Futures string // "none" if nothing
}

var quotePref = []string{"USDT", "USDC", "USD", "EUR", "KRW", "BTC"}

// func canonPair(base, quote string) string {
// 	return strings.ToUpper(base) + "/" + strings.ToUpper(quote)
// }

// Index of the presence of the coin on the target exchange (по spot)
func BuildPresence(items []dm.Item) map[string]Presence {
	m := make(map[string]Presence)
	for _, it := range items {
		if it.Type != dm.TypeSpot {
			continue
		}
		base := strings.ToUpper(it.Base)
		quote := strings.ToUpper(it.Quote)
		pr := m[base]
		if pr.Quotes == nil {
			pr.Quotes = make(map[string]struct{})
		}
		pr.Quotes[quote] = struct{}{}
		pr.HasAny = true
		if quote == "BTC" {
			pr.HasBTC = true
		} else {
			pr.HasNonBTC = true
		}
		m[base] = pr
	}
	return m
}

// Index by source: choose 1 best spot ticker and (if any) futures
func BuildSourceIndex(items []dm.Item) map[string]SourceInfo {
	// base -> (quote->symbol)
	spot := make(map[string]map[string]string)
	futs := make(map[string]string)

	for _, it := range items {
		base := strings.ToUpper(it.Base)
		switch it.Type {
		case dm.TypeSpot:
			if spot[base] == nil {
				spot[base] = make(map[string]string)
			}
			spot[base][strings.ToUpper(it.Quote)] = it.Symbol
		case dm.TypeFutures:
			// we store “any” futures symbol for this coin
			if _, ok := futs[base]; !ok {
				futs[base] = it.Symbol
			}
		}
	}

	out := make(map[string]SourceInfo, len(spot))
	for base, q2sym := range spot {
		var chosen string
		// choose the best spot ticker by quote priority
		for _, q := range quotePref {
			if s, ok := q2sym[q]; ok {
				chosen = s
				break
			}
		}
		if chosen == "" { // if there is no priority, we take any
			for _, s := range q2sym { chosen = s; break }
		}
		var fs *string
		if sym, ok := futs[base]; ok && sym != "" {
			fs = &sym
		}
		out[base] = SourceInfo{SpotSymbol: chosen, FuturesSymbol: fs}
	}
	return out
}

// Build list lines according to goal rules.
// mode: "upbit" / "bithumb" (ignore BTC-only) or "coinbase" (exclude any presence)
func MakeList(src map[string]SourceInfo, target map[string]Presence, mode string) []Row {
	rows := make([]Row, 0, len(src))
	for base, si := range src {
		pr := target[base]
		exclude := false
		switch mode {
		case "upbit", "bithumb":
			// we exclude only if there is a NON-BTC quote on the target
			if pr.HasNonBTC {
				exclude = true
			}
		case "coinbase":
			// exclude any presence
			if pr.HasAny {
				exclude = true
			}
		default:
			// default as coinbase
			if pr.HasAny {
				exclude = true
			}
		}
		if exclude {
			continue
		}
		f := "none"
		if si.FuturesSymbol != nil && *si.FuturesSymbol != "" {
			f = *si.FuturesSymbol
		}
		rows = append(rows, Row{Spot: si.SpotSymbol, Futures: f})
	}

	// sort by spot
	sort.Slice(rows, func(i, j int) bool { return rows[i].Spot < rows[j].Spot })
	return rows
}
