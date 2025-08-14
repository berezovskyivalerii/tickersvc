package markets

type Type string

const (
	TypeSpot    Type = "spot"
	TypeFutures Type = "futures"
)

type Item struct {
	ExchangeID   int16
	Type         Type
	Symbol       string
	Base         string
	Quote        string
	ContractSize *int64 // nil for spot
	Active       bool   // true by defaut
}
