package api

import (
	"time"
)

// TimeResponse wraps our time information to be returned as a JSON response
// against the `/time` endpoint. By default, we return the time in UTC, but can
// change this behavior via the `tz` query paremeter. Valid values for `tz` are
// dictated by IANA's time zone database.
//
// See:
// - https://www.iana.org/time-zones
// - https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
//
// `Now` is set to the time the server's handler began processing the request.
// `Zone` is the abbreviated timezone, e.g. EST, CET, UTC, etc.
// `Offset` is the time offset in seconds (east of UTC) for the specified zone.
// `UTC` is our time in UTC.
// `Unix` is the number of seconds elapsed since 01/01/1970 UTC.
// `Unix` is the number of nanoseconds elapsed since 01/01/1970 UTC.
type TimeResponse struct {
	Now      time.Time `json:"now"`
	Zone     string    `json:"zone"`
	Offset   int       `json:"offset"`
	UTC      time.Time `json:"utc"`
	Unix     int64     `json:"unix"`
	UnixNano int64     `json:"unixNano"`
}
