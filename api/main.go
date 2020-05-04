package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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

var bind string
var failureRate float64
var delayInterval string
var minDelay int
var maxDelay int

func init() {
	flag.StringVar(&bind, "bind", ":8080", "the bind address and port for the server to listen on")
	flag.Float64Var(&failureRate, "failure-rate", 0.10, "percentage of requests to respond with 500s to, e.g. 0.13 for 13%")
	flag.StringVar(&delayInterval, "delay-interval", "0,100", "add a random delay in milliseconds before processing a request")
	flag.Parse()

	interval := strings.Split(delayInterval, ",")
	if len(interval) != 2 {
		log.Fatal("Interval must be two integers delimited by a comma")
	}

	var err error

	if minDelay, err = strconv.Atoi(interval[0]); err != nil {
		log.Fatal("Failed parsing min delay as an integer")
	}

	if maxDelay, err = strconv.Atoi(interval[1]); err != nil {
		log.Fatal("Failed parsing max delay as an integer")
	}
}

func sleepRandomly(min, max int) {
	duration := time.Duration(rand.Intn(max-min+1)+min) * time.Millisecond
	time.Sleep(duration)
}

func main() {
	rand.Seed(time.Now().Unix())

	http.HandleFunc("/time", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		sleepRandomly(minDelay, maxDelay)
		if failureRate > rand.Float64() {
			http.Error(w, "oops random failure", 500)
			return
		}

		tz := "UTC"
		tzs, _ := r.URL.Query()["tz"]
		if len(tzs) > 0 {
			tz = tzs[0]
		}

		location, err := time.LoadLocation(tz)
		if err != nil {
			http.Error(w, "Bad timezone!", 400)
			return
		}

		now = now.In(location)
		zone, offset := now.Zone()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&TimeResponse{
			Now:      now,
			Zone:     zone,
			Offset:   offset,
			UTC:      now.UTC(),
			Unix:     now.Unix(),
			UnixNano: now.UnixNano(),
		})
	})

	log.Printf("Listening on %s", bind)
	if err := http.ListenAndServe(bind, nil); err != nil {
		log.Fatal(err)
	}
}
