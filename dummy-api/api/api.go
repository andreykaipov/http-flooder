package api

import (
	"time"
)

// TimeResponse wraps our time information to be returned as a JSON response
// against the `/time` endpoint. By default, we return the time in UTC, but can
// change this behavior via the `tz` query paremeter. Valid values for `tz` are
// dictated by IANA's time zone database.
//
// See https://www.iana.org/time-zones and
// https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
type TimeResponse struct {

	// Now is set to the time the server's handler began processing the
	// request.
	Now      time.Time `json:"now"`

	// Zone is the abbreviated timezone, e.g. EST, CET, UTC, etc.
	Zone     string    `json:"zone"`

	// Offset is the time offset in seconds (east of UTC) for the specified
	// zone.
	Offset   int       `json:"offset"`

	// UTC is ourtime in UTC.
	UTC      time.Time `json:"utc"`

	// Unix is the number of seconds elapsed since 01/01/1970 UTC.
	Unix     int64     `json:"unix"`

	// UnixNano is the number of nanoseconds elapsed since 01/01/1970 UTC.
	UnixNano int64     `json:"unixNano"`

}
