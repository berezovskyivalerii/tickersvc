package healthjson

import (
	"net/http"

	domain "github.com/berezovskyivalerii/tickersvc/internal/domain/health"
	usecase "github.com/berezovskyivalerii/tickersvc/internal/usecase/health"
)

type Response struct {
	Status    string            `json:"status"`
	Version   string            `json:"version,omitempty"`
	Commit    string            `json:"commit,omitempty"`
	BuildTime string            `json:"buildTime,omitempty"`
	Uptime    string            `json:"uptime,omitempty"`
	Checks    map[string]string `json:"checks"`
	Now       string            `json:"now,omitempty"`
}

func Map(out usecase.ReadinessOutput) (int, Response) {
	code := http.StatusOK
	if out.Status == domain.StatusDegraded {
		code = http.StatusServiceUnavailable
	}
	resp := Response{
		Status:    string(out.Status),
		Version:   out.Version,
		Commit:    out.Commit,
		BuildTime: out.BuildTime,
		Uptime:    out.Uptime.String(),
		Checks:    map[string]string{},
		Now:       out.Now.Format("2006-01-02T15:04:05Z07:00"),
	}
	for k, v := range out.Checks {
		resp.Checks[k] = string(v)
	}
	return code, resp
}
