//go:build !exclude_swagger
// +build !exclude_swagger

package app

// Общие модели ответов для примеров в Swagger

type HealthResp struct {
	Status   string            `json:"status"`
	Version  string            `json:"version"`
	Commit   string            `json:"commit"`
	BuildTime string           `json:"buildTime"`
	Uptime   string            `json:"uptime"`
	Checks   map[string]string `json:"checks"`
	Now      string            `json:"now"`
}

type UpdateResp struct {
	MarketsSync     map[string][3]int `json:"markets_sync"`
	SegmentsUpdated map[string]int    `json:"segments_updated,omitempty"`
	ListsUpdated    map[string]int    `json:"lists_updated,omitempty"`
}

type ListGetJSON struct {
	Slug  string   `json:"slug"`
	Items []string `json:"items"`
}

// @title TickerSvc API
// @version 1.0
// @BasePath /
// @schemes http
// @description API для сегментов (seg1..seg4) и таргет-списков.
// @contact.name API Support

// Health
// @Summary     Service readiness
// @Tags        public
// @Produce     json
// @Success     200 {object} HealthResp
// @Router      /health [get]
func _doc_health() {}

// Update
// @Summary     Sync markets and rebuild lists
// @Tags        admin
// @Param       mode   query  string false "segments|targets|all"
// @Param       source query  string false "binance,bybit,okx"
// @Param       target query  string false "upbit|coinbase|bithumb"
// @Produce     json
// @Success     200 {object} UpdateResp
// @Failure     500 {object} map[string]string
// @Router      /update [post]
func _doc_update() {}

// Lists by slug
// @Summary     Get list by slug
// @Tags        public
// @Param       slug     path   string true  "list slug" Example(binance_seg1)
// @Param       as_text  query  int    false "1 → text/plain"
// @Produce     json
// @Success     200 {object} ListGetJSON
// @Router      /api/lists/{slug} [get]
func _doc_lists() {}

// Segments sugar-redirect
// @Summary     Segment view (redirects to /api/lists/{source}_seg{n})
// @Tags        public
// @Param       source path string true "binance|bybit|okx"
// @Param       seg    path int    true "1|2|3|4"
// @Param       as_text query int  false "1 → text/plain"
// @Success     307 {string} string "Temporary Redirect"
// @Router      /api/segments/{source}/{seg} [get]
func _doc_segments() {}
