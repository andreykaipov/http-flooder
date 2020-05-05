package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"os"
	"sync"
	"time"
)

var (
	endpoint          string
	requestsPerSecond int
	duration          int
	timeout           int
	concurrency       int
	maxRetry          int
	report            string
	verbose           bool
)

// TODO: support maxRetry flag
//
// Whenever a request times out or the server "mysteriously" responds with an
// error, we should (optionally) retry the request to make sure it got through.
// Note this will raise the total request count (so that request total is no
// longer success + failues). In addition, we should probably distinguish
// between client and server errors. That's good, but the actual tricky part is
// requeuing the request. We could make a `requests` channel and use that to
// queue outgoing requests, at which point requeuing would involve just putting
// it back onto the channel. Each request should also know the current attempt
// it's on, as that's what we can use to compare against our `maxRetry` flag.

func init() {
	flag.StringVar(&endpoint, "endpoint", "", "the endpoint to GET, e.g. http://cool-api:8080/wow")
	flag.IntVar(&requestsPerSecond, "requests-per-second", 1, "number of GET requests per second to initiate against the endpoint")
	flag.IntVar(&duration, "duration", 10, "how long would we like to run the test for (in seconds)?")
	flag.IntVar(&timeout, "timeout", 1000, "timeout in milliseconds for the request to finish before failing")
	flag.IntVar(&maxRetry, "maxRetry", 3, "max retries to make for failed requests (unused)")
	flag.StringVar(&report, "report", "", "output the report to a JSON file")
	flag.BoolVar(&verbose, "verbose", false, "verbose logging while querying the servers")
	flag.Parse()
}

func main() {
	if endpoint == "" {
		fmt.Fprintln(os.Stderr, "You gotta specify an endpoint.")
		os.Exit(1)
	}

	// Disabling keepalives on our client's transport is important so we can
	// capture the start of the connection (i.e. httptrace's `ConnectStart`
	// event). Otherwise, since our host machine keeps the connection open
	// to the server after our request ends, any subsequent requests will
	// not need to initiate a connection, and we'll be unable to calculate
	// the TTFB and TTLB accurately for them.
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	agg := &Aggregation{}
	jobs := sync.WaitGroup{}
	secondsTicker := time.NewTicker(1 * time.Second)
	tick := 0

	fmt.Printf(`Starting flood. :-)
Running for %d second(s), initiating %d request(s) per second. Total requests send to server will be %d.
`,
		duration,
		requestsPerSecond,
		duration*requestsPerSecond,
	)
flood:
	for {
		select {
		case <-secondsTicker.C:
			fmt.Println("Sending batch", tick)

			tick++
			if tick > duration {
				secondsTicker.Stop()
				break flood
			}

			go func() {
				jobs.Add(1)
				for i := 0; i < requestsPerSecond; i++ {
					go func() {
						jobs.Add(1)
						go get(client, endpoint, agg)
						jobs.Done()
					}()
				}
				jobs.Done()
			}()
		}
	}

	jobs.Wait()

	agg.PrettyPrint()
	if report != "" {
		agg.Write(report)
	}
}

// get implements the HTTP tracing logic to calculate TTFB/TTLB, while also
// adding these statistics to our Aggregation object.
//
// see https://golang.org/pkg/net/http/httptrace/#ClientTrace for the builtin
// events the httptrace package exposes for client HTTP requests.
func get(client *http.Client, endpoint string, agg *Aggregation) {
	var connectStart, firstByte, bodyRead time.Time

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		agg.AddFailure()
		fmt.Fprintf(os.Stderr, "Failed forming request %v: %v\n", endpoint, err)
		return
	}
	req.Close = true

	resp, err := client.Do(req.WithContext(httptrace.WithClientTrace(
		req.Context(),
		&httptrace.ClientTrace{
			ConnectStart:         func(_, _ string) { connectStart = time.Now() },
			GotFirstResponseByte: func() { firstByte = time.Now() },
		},
	)))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed issuing request %v: %v\n", endpoint, err)
		return
	}

	if resp.StatusCode/100 != 2 {
		agg.AddFailure()
		fmt.Fprintf(os.Stderr, "Received non-2xx response from server: %v\n", resp.StatusCode)
		return
	}

	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		agg.AddFailure()
		fmt.Fprintf(os.Stderr, "Failed reading response %v: %v\n", resp, err)
		return
	}

	// The body was read. That was the last byte.
	bodyRead = time.Now()

	if err := resp.Body.Close(); err != nil {
		agg.AddFailure()
		fmt.Fprintf(os.Stderr, "Failed closing response body %v\n", err)
		return
	}

	ttfb, ttlb := firstByte.Sub(connectStart), bodyRead.Sub(connectStart)

	agg.AddSuccess(ttfb, ttlb)
	if verbose {
		fmt.Printf("ttfb=%s ttlb=%s delta=%s\n", ttfb, ttlb, ttlb-ttfb)
	}
}
