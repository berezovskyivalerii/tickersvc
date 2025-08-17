package lists

import (
	"fmt"
	"sort"

	listsdom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
)

type SegmentKind string

const (
	Seg1 SegmentKind = "seg1" // 🟢 S \ U
	Seg2 SegmentKind = "seg2" // 🟠 S ∩ ((U∩C) ∪ (C∩H))
	Seg3 SegmentKind = "seg3" // 🔴 S ∩ U ∩ C ∩ H
	Seg4 SegmentKind = "seg4" // 🔵 S ∩ (U \ (C ∪ H))
)

type Segments struct {
	Seg1 []listsdom.Row
	Seg2 []listsdom.Row
	Seg3 []listsdom.Row
	Seg4 []listsdom.Row
}

func BuildSegmentsForSource(sets Sets, source string) (Segments, error) {
	var S map[string]SourceInf
	switch source {
	case "binance":
		S = sets.Binance
	case "bybit":
		S = sets.Bybit
	case "okx":
		S = sets.OKX
	default:
		return Segments{}, fmt.Errorf("unknown source %q", source)
	}

	inU, inH, inC := sets.Upbit, sets.Bithumb, sets.Coinbase
	in := func(m map[string]struct{}, k string) bool { _, ok := m[k]; return ok }

	seg3, seg4, seg2, seg1 := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}

	// 🔴 S ∩ U ∩ C ∩ H
	for base := range S {
		if in(inU, base) && in(inC, base) && in(inH, base) {
			seg3[base] = struct{}{}
		}
	}
	// 🔵 S ∩ (U \ (C ∪ H))
	for base := range S {
		if in(inU, base) && !in(inC, base) && !in(inH, base) {
			if _, taken := seg3[base]; !taken {
				seg4[base] = struct{}{}
			}
		}
	}
	// 🟠 S ∩ ((U∩C) ∪ (C∩H)) \ (seg3 ∪ seg4)
	for base := range S {
		uc := in(inU, base) && in(inC, base)
		ch := in(inC, base) && in(inH, base)
		if uc || ch {
			if _, red := seg3[base]; red {
				continue
			}
			if _, blue := seg4[base]; blue {
				continue
			}
			seg2[base] = struct{}{}
		}
	}
	// 🟢 S \ U  (и не дублируем то, что уже пошло в 🟠)
	for base := range S {
		if !in(inU, base) {
			if _, orange := seg2[base]; orange {
				continue
			}
			seg1[base] = struct{}{}
		}
	}

	toRows := func(bases map[string]struct{}) []listsdom.Row {
		out := make([]listsdom.Row, 0, len(bases))
		for base := range bases {
			si := S[base]
			var fut *string
			if si.FuturesSymbol != "" {
				f := si.FuturesSymbol
				fut = &f
			}
			out = append(out, listsdom.Row{Spot: si.SpotSymbol, Futures: fut})
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Spot < out[j].Spot })
		return out
	}

	return Segments{
		Seg1: toRows(seg1),
		Seg2: toRows(seg2),
		Seg3: toRows(seg3),
		Seg4: toRows(seg4),
	}, nil
}

func BuildAllSegments(sets Sets) map[string][]listsdom.Row {
	out := make(map[string][]listsdom.Row, 12)

	if segs, err := BuildSegmentsForSource(sets, "binance"); err == nil {
		out["binance_seg1"] = segs.Seg1
		out["binance_seg2"] = segs.Seg2
		out["binance_seg3"] = segs.Seg3
		out["binance_seg4"] = segs.Seg4
	}
	if segs, err := BuildSegmentsForSource(sets, "bybit"); err == nil {
		out["bybit_seg1"] = segs.Seg1
		out["bybit_seg2"] = segs.Seg2
		out["bybit_seg3"] = segs.Seg3
		out["bybit_seg4"] = segs.Seg4
	}
	if segs, err := BuildSegmentsForSource(sets, "okx"); err == nil {
		out["okx_seg1"] = segs.Seg1
		out["okx_seg2"] = segs.Seg2
		out["okx_seg3"] = segs.Seg3
		out["okx_seg4"] = segs.Seg4
	}

	return out
}
