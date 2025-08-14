package lists

import (
	"sort"
	"strings"

	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

// BuildListRows builds a list according to the rules of the TOR.
// - source: all source instruments (spot+futures)
// - target: all target exchange instruments (only spot is considered for presence)
// - ignoreBTCOnly: true for Upbit/Bithumb (BTC-only on target is NOT considered presence)
func BuildListRows(source []dm.Item, target []dm.Item, mode string) []Row {
	srcIdx := buildSourceIndex(source)
	tgtPr  := buildPresence(target)

	rows := make([]Row, 0, len(srcIdx))
	for base, si := range srcIdx {
		pr := tgtPr[base]
		exclude := false

		switch mode {
		case "upbit", "bithumb":
			// we exclude only if there is a NON-BTC spot on the target
			exclude = pr.HasNonBTC
		case "coinbase":
			// we exclude at any spot presence
			exclude = pr.HasAnySpot
		case "binance":
			// we exclude only if there is spot and no futures on Binance
			exclude = pr.HasAnySpot && !pr.HasFutures
		default:
			exclude = pr.HasAnySpot
		}
		if exclude { continue }

		f := "none"
		if si.FuturesSymbol != "" { f = si.FuturesSymbol }
		rows = append(rows, Row{Spot: si.SpotSymbol, Futures: f})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Spot < rows[j].Spot })
	return rows
}


// FormatAsText — "SPOT, FUTURES" line by line.
func FormatAsText(rows []Row) string {
	if len(rows) == 0 {
		return ""
	}
	var b strings.Builder
	for i, r := range rows {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(r.Spot)
		b.WriteString(", ")
		b.WriteString(r.Futures)
	}
	b.WriteByte('\n')
	return b.String()
}

// ===== AUXILIARY (local, no export) =====
type presence struct {
	HasAnySpot bool
	HasBTC     bool
	HasNonBTC  bool
	HasFutures bool
}

func buildPresence(items []dm.Item) map[string]presence {
	m := make(map[string]presence)
	for _, it := range items {
		base := strings.ToUpper(it.Base)
		pr := m[base]
		switch it.Type {
		case dm.TypeSpot:
			pr.HasAnySpot = true
			q := strings.ToUpper(it.Quote)
			if q == "BTC" { pr.HasBTC = true } else { pr.HasNonBTC = true }
		case dm.TypeFutures:
			pr.HasFutures = true
		}
		m[base] = pr
	}
	return m
}

type srcInfo struct {
	SpotSymbol    string
	FuturesSymbol string // "" если нет
}

func buildSourceIndex(items []dm.Item) map[string]srcInfo {
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
			// save the first futures ticker for the coin
			if _, ok := futs[base]; !ok && it.Symbol != "" {
				futs[base] = it.Symbol
			}
		}
	}

	// priority of quoted currencies to select the "best" spot
	quotePref := []string{"USDT", "USDC", "USD", "EUR", "KRW", "BTC"}

	out := make(map[string]srcInfo, len(spot))
	for base, q2s := range spot {
		var chosen string
		for _, q := range quotePref {
			if s, ok := q2s[q]; ok {
				chosen = s
				break
			}
		}
		// if we didn't find it by priority, we'll take any
		if chosen == "" {
			for _, s := range q2s {
				chosen = s
				break
			}
		}
		out[base] = srcInfo{
			SpotSymbol:    chosen,
			FuturesSymbol: futs[base], // can be ""
		}
	}
	return out
}
