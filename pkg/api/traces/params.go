package traces

import (
	"net/url"
	"strconv"
	"time"
)

type UrlQueryParams struct {
	Start time.Time
	End   time.Time
	Limit int
}

func (x *UrlQueryParams) ReadParams(params url.Values) {
	x.Start, _ = time.Parse(time.RFC3339, params.Get("start"))
	if x.Start.IsZero() {
		x.Start = time.Now().Add(-52 * 7 * 24 * 3600 * time.Second)
	}

	x.End, _ = time.Parse(time.RFC3339, params.Get("end"))
	if x.End.IsZero() {
		x.End = time.Now()
	}

	x.Limit, _ = strconv.Atoi(params.Get("limit"))
	if x.Limit == 0 {
		x.Limit = 1
	}
}
