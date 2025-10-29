package lists

import (
	"context"
	"strings"

	"github.com/berezovskyivalerii/tickersvc/internal/config"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ( 
	ExBinance int16 = 1 
	ExBybit int16 = 2 
	ExOKX int16 = 3 
	ExCoinbase int16 = 4 
	ExUpbit int16 = 5 
	ExBithumb int16 = 6 
	)

type SourceInf struct { 
	Base string 
	SpotSymbol string 
	FuturesSymbol string 
}

// конвертация SourceInfo -> SourceInf (string вместо *string)
func adaptInfo(in map[string]SourceInfo) map[string]SourceInf {
    out := make(map[string]SourceInf, len(in))
    for base, si := range in {
        f := ""
        if si.FuturesSymbol != nil {
            f = *si.FuturesSymbol
        }
        out[base] = SourceInf{
            Base:          base,
            SpotSymbol:    si.SpotSymbol,
            FuturesSymbol: f,
        }
    }
    return out
}

type Sets struct {
	Binance map[string]SourceInf
	Bybit   map[string]SourceInf
	OKX     map[string]SourceInf

	Upbit    map[string]struct{}
	Bithumb  map[string]struct{}
	Coinbase map[string]struct{}
}


func buildPresenceFiltered(items []dm.Item, ignoreQuotes map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{})
	for _, it := range items {
		if it.Type != dm.TypeSpot {
			continue
		}
		q := strings.ToUpper(it.Quote)
		if _, skip := ignoreQuotes[q]; skip {
			continue
		}
		base := strings.ToUpper(it.Base)
		out[base] = struct{}{}
	}
	return out
}

type RawByEx struct {
	Binance  []dm.Item
	Bybit    []dm.Item
	OKX      []dm.Item
	Upbit    []dm.Item
	Bithumb  []dm.Item
	Coinbase []dm.Item
}

func BuildSetsFromRaw(raw RawByEx) Sets {
    ignoreUH := map[string]struct{}{"BTC": {}, "USDT": {}}

    return Sets{
        Binance:  adaptInfo(BuildSourceIndex(raw.Binance)),
        Bybit:    adaptInfo(BuildSourceIndex(raw.Bybit)),
        OKX:      adaptInfo(BuildSourceIndex(raw.OKX)),

        Upbit:    buildPresenceFiltered(raw.Upbit,    ignoreUH),
        Bithumb:  buildPresenceFiltered(raw.Bithumb,  ignoreUH),
        Coinbase: buildPresenceFiltered(raw.Coinbase, nil), // без фильтра
    }
}

func BuildSets(ctx context.Context, mr dm.Repo, quotes config.QuotesConfig) (Sets, error) { 
	load := func(ex int16) ([]dm.Item, error) { 
		return mr.LoadActiveByExchange(ctx, ex) 
	} 
	bin, err := load(ExBinance) 
	if err != nil { return Sets{}, err } 
	byb, err := load(ExBybit) 
	if err != nil { return Sets{}, err } 
	okx, err := load(ExOKX) 
	if err != nil { return Sets{}, err } 
	upb, err := load(ExUpbit) 
	if err != nil { return Sets{}, err } 
	bth, err := load(ExBithumb) 
	if err != nil { return Sets{}, err } 
	cnb, err := load(ExCoinbase) 
	if err != nil { return Sets{}, err } 
	out := Sets{ Binance: map[string]SourceInf{}, 
				 Bybit: map[string]SourceInf{}, 
				 OKX: map[string]SourceInf{}, 
				 Upbit: map[string]struct{}{}, 
				 Bithumb: map[string]struct{}{}, 
				 Coinbase: map[string]struct{}{}, 
	} 
	srcQuote := strings.ToUpper(quotes.SourceSpotQuote) 
	type futMap = map[string]string 
	futB, futY, futO := futMap{}, futMap{}, futMap{} 
	// --- источники --- 
	handleSource := func(items []dm.Item, ex int16, spot map[string]SourceInf, futs futMap) { 
		for _, it := range items { 
			base := strings.ToUpper(it.Base) 
			quote := strings.ToUpper(it.Quote) 
			switch it.Type { 
				case dm.TypeSpot: 
					if quote != srcQuote { continue } 
					if _, ok := spot[base]; !ok { 
						spot[base] = SourceInf{Base: base, SpotSymbol: it.Symbol} 
					} 
				case dm.TypeFutures: 
					if _, ok := futs[base]; !ok {
						futs[base] = it.Symbol 
					} 
				} 
			} 
	} 

	handleSource(bin, ExBinance, out.Binance, futB) 

	handleSource(byb, ExBybit, out.Bybit, futY) 
	handleSource(okx, ExOKX, out.OKX, futO) 

	for base, si := range out.Binance { 
		if f, ok := futB[base]; ok { si.FuturesSymbol = f 
			out.Binance[base] = si } 
	} 
	for base, si := range out.Bybit { 
		if f, ok := futY[base]; ok { si.FuturesSymbol = f 
		out.Bybit[base] = si } 
	} 
	for base, si := range out.OKX { 
		if f, ok := futO[base]; ok { 
			si.FuturesSymbol = f 
			out.OKX[base] = si 
		} 
	} 
	isAllowed := func(q string) bool { _, ok := quotes.TargetAllowedQuotes[q] 
		return ok } 
	handleTarget := func(items []dm.Item, dst map[string]struct{}, allow func(q string) bool) { 
		for _, it := range items { 
			if it.Type != dm.TypeSpot { continue } 
			q := strings.ToUpper(it.Quote) 
			if !allow(q) { continue }
			base := strings.ToUpper(it.Base) 
			dst[base] = struct{}{} } 
		} 
	allowUpOrBithumb := func(q string) bool { 
		if !isAllowed(q) { return false } 
		return q != "USDT" && q != "BTC" 
		} 
		allowCoinbase := func(q string) bool { return isAllowed(q) } 
		handleTarget(upb, out.Upbit, allowUpOrBithumb) 
		handleTarget(bth, out.Bithumb, allowUpOrBithumb) 
		handleTarget(cnb, out.Coinbase, allowCoinbase) 
	return out, nil 
}


