package marketsuc

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

func ansi(code string, s string) string { return "\x1b[" + code + "m" + s + "\x1b[0m" }
// helpers
func green(s string) string { return ansi("32", s) }
func yellow(s string) string { return ansi("33", s) }
func red(s string) string    { return ansi("31", s) }
func dim(s string) string    { return ansi("2", s) }

// FormatSummary печатает таблицу: EXCHANGE | ADDED | UPDATED | ARCHIVED
// id2name: карта ID биржи -> slug/name (для читаемости)
func FormatSummary(sum map[int16][3]int, id2name map[int16]string) string {
	// сортируем по имени
	type row struct{ name string; a, u, d int }
	rows := make([]row, 0, len(sum))
	var ta, tu, td int
	for id, v := range sum {
		name := id2name[id]
		if name == "" { name = strconv.Itoa(int(id)) }
		rows = append(rows, row{name, v[0], v[1], v[2]})
		ta += v[0]; tu += v[1]; td += v[2]
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].name < rows[j].name })

	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "EXCHANGE\tADDED\tUPDATED\tARCHIVED")
	for _, r := range rows {
		a := fmt.Sprint(r.a)
		u := fmt.Sprint(r.u)
		d := fmt.Sprint(r.d)
		if r.a > 0 { a = green(a) }
		if r.u > 0 { u = yellow(u) }
		if r.d > 0 { d = red(d) }
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.name, a, u, d)
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", dim("TOTAL"), dim(fmt.Sprint(ta)), dim(fmt.Sprint(tu)), dim(fmt.Sprint(td)))
	_ = w.Flush()
	return b.String()
}
