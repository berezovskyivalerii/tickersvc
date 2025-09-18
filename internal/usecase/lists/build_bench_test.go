package lists

import (
	"testing"

	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

func BenchmarkBuildListRows(b *testing.B) {
	var src []dm.Item
	for i:=0; i<5000; i++ {
		base := "T"+itoa(i)
		src = append(src, dm.Item{Type:dm.TypeSpot, Base:base, Quote:"USDT", Symbol: base+"USDT", Active:true})
	}
	for i:=0; i<2000; i++ {
		base := "T"+itoa(i)
		src = append(src, dm.Item{Type:dm.TypeFutures, Base:base, Quote:"USDT", Symbol: base+"USDT-PERP", Active:true})
	}
	var tgt []dm.Item
	for i:=0; i<2000; i++ {
		base := "T"+itoa(i)
		tgt = append(tgt, dm.Item{Type:dm.TypeSpot, Base:base, Quote:"USDT", Symbol:"USDT-"+base, Active:true})
	}

	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		_ = BuildListRows(src, tgt, "upbit")
	}
}

func itoa(i int) string {
	if i == 0 { return "0" }
	buf := [20]byte{}
	pos := len(buf)
	n := i
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n/=10
	}
	return string(buf[pos:])
}
