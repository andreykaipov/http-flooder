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

	api "./api"
)

var (
	bind          string
	failureRate   float64
	delayInterval string
	minDelay      int
	maxDelay      int
)

func init() {
	flag.StringVar(&bind, "bind", ":8080", "the bind address and port for the server to listen on")
	flag.Float64Var(&failureRate, "failureRate", 0.10, "percentage of requests to respond with 500s to, e.g. 0.13 for 13%")
	flag.StringVar(&delayInterval, "delayInterval", "0,100", "add a random delay in milliseconds before processing a request")
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
		tzs := r.URL.Query()["tz"]
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
		if err := json.NewEncoder(w).Encode(&api.TimeResponse{
			Now:      now,
			Zone:     zone,
			Offset:   offset,
			UTC:      now.UTC(),
			Unix:     now.Unix(),
			UnixNano: now.UnixNano(),
		}); err != nil {
			http.Error(w, "We've failed you. :(", 500)
		}
	})

	log.Printf("Listening on %s", bind)

	if err := http.ListenAndServe(bind, nil); err != nil {
		log.Fatal(err)
	}
}
