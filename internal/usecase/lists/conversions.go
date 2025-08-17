package lists

import listsdom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"

// FromDomainRows переводит доменные строки (Futures *string)
// в устаревший формат этого пакета (Futures string с "none").
func FromDomainRows(in []listsdom.Row) []Row {
	out := make([]Row, 0, len(in))
	for _, r := range in {
		fut := "none"
		if r.Futures != nil && *r.Futures != "" {
			fut = *r.Futures
		}
		out = append(out, Row{Spot: r.Spot, Futures: fut})
	}
	return out
}
