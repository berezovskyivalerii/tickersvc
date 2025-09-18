package lists

import (
	"fmt"
	"sort"

	listsdom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
)

type SegmentKind string

const (
	Seg0 SegmentKind = "seg0" // ◯ S \ (U ∪ H ∪ C) — используем только для binance
	Seg1 SegmentKind = "seg1" // 🟢 S \ U
	Seg2 SegmentKind = "seg2" // 🟠 S ∩ ((U∩C) ∪ (C∩H))
	Seg3 SegmentKind = "seg3" // 🔴 S ∩ U ∩ C ∩ H
	Seg4 SegmentKind = "seg4" // 🔵 S ∩ (U \ (C ∪ H))
)

type Segments struct {
	Seg0 []listsdom.Row
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

    seg0 := map[string]struct{}{} // ◯ S \ (U ∪ H ∪ C) — только binance
    seg1 := map[string]struct{}{} // 🟢 S ∩ ((C ∪ H) \ U)
    seg2 := map[string]struct{}{} // 🟠 S ∩ ( ((U∩H)\C) ∪ ((U∩C)\H) )
    seg3 := map[string]struct{}{} // 🔴 S ∩ U ∩ C ∩ H
    seg4 := map[string]struct{}{} // 🔵 S ∩ (U \ (C ∪ H))

    for base := range S {
        u, h, c := in(inU, base), in(inH, base), in(inC, base)

        if source == "binance" && !u && !h && !c {
            seg0[base] = struct{}{}
        }
        if u && c && h {
            seg3[base] = struct{}{}
            continue
        }
        if u && !c && !h {
            seg4[base] = struct{}{}
            continue
        }
        // 🟢: присутствует на C или H, но отсутствует на Up
        if (c || h) && !u {
            seg1[base] = struct{}{}
            continue
        }
        // 🟠: на Up вместе ровно с одной из бирж (вторая отсутствует)
        if u && ((h && !c) || (c && !h)) {
            seg2[base] = struct{}{}
            continue
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
        Seg0: toRows(seg0),
        Seg1: toRows(seg1),
        Seg2: toRows(seg2),
        Seg3: toRows(seg3),
        Seg4: toRows(seg4),
    }, nil
}

func BuildAllSegments(sets Sets) map[string][]listsdom.Row {
	out := make(map[string][]listsdom.Row, 12)

	if segs, err := BuildSegmentsForSource(sets, "binance"); err == nil {
		out["binance_seg0"] = segs.Seg0
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
