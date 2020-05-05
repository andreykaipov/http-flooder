# HTTP Flooder

Hello! This project contains:

1. A CLI to flood HTTP servers with requests (in the `flooder` directory).
1. A dummy api (in the `dummy-api` directory).

Table of Contents:
- [Flooder](#flooder) - documentation for the `flooder` CLI.
- [Dummy API](#dummy-api) - documentation for the dummy API.
- [Docker](#docker) - examples showcasing the interaction between the two via
  Docker.

Both applications are written in Go, so you'll need to have it installed to be
able to compile the programs. They only use the standard library, so don't
worry - mucking around with the GOPATH or using Go modules isn't necessary.

You can also skip to the [Docker](#docker) section to just build the
applications with that.

## Flooder

### How does it work?

The `flooder` initiates concurrent HTTP requests against a specified server
endpoint. We can customize the way it works through the following flags:

| flag                | type   | description                                                        | default |
|---------------------|--------|--------------------------------------------------------------------|---------|
| duration            | int    | how long would we like to run the test for (in seconds)?           | `10`    |
| endpoint            | string | the endpoint to GET, e.g. http://cool-api:8080/wow                 |         |
| maxRetry            | int    | max retries to make for failed requests **(not implemented)**      | `3`     |
| report              | string | output the report to a JSON file                                   | `""`    |
| requestsPerSecond   | int    | number of GET requests per second to initiate against the endpoint | `1`     |
| timeout             | int    | timeout in milliseconds for the request to finish before failing   | `1000`  |
| verbose             | bool   | verbose logging while querying the servers                         | `false` |

After a run, the tool will output a summary about how well the server performed
against our flood of requests.

### Let's see

```console
❯ make flooder

❯ ./bin/flooder -endpoint http://google.com -requestsPerSecond 3 -duration 4 -report report.json
Starting flood. :-)
Running for 4 second(s), initiating 3 request(s) per second. Total requests send to server will be 12.
Sending batch 1
Sending batch 2
Sending batch 3
Sending batch 4

Total Requests: 12
     Successes: 12
      Failures: 0
  Success Rate: 100.00000%
  Failure Rate: 0.00000%
  Average TTFB: 130.582007ms
  Average TTLB: 131.13785ms
         Delta: 555.843µs

❯ jq . report.json
{
  "successes": 12,
  "failures": 0,
  "ttfb": 130582007,
  "ttlb": 131137850
}
```

The definitions for TTFB and TTLB that we're using are as follows:

**Time to first byte** - the duration between the time when we initiate
a connection to the server and the time when we receive the first byte of the
server's response headers.

**Time to last byte** - the duration between the time when we initiate
a connection to the server and the time when we finish reading the server's
response body.

## Dummy API

### How does it work?

A dummy API is exactly what it sounds like. It'll spin up a web server serving
one endpoint `/time` that will output time information about whatever timezone
we pass it via the `tz` query parameter (UTC by default). See examples below.

In addition, we can customize how reliable our web server behaves through the
following flags, namely `delayInterval` and `failureRate`.

| flag           | type   | description                                                       | default   |
|----------------|--------|-------------------------------------------------------------------|-----------|
| bind           | string | the bind address and port for the server to listen on             | `":8080"` |
| delayInterval  | string | add a random delay in milliseconds before processing a request    | `"0,100"` |
| failureRate    | float  | percentage of requests to respond with 500s to, e.g. 0.13 for 13% | `0.1`     |

### Let's see

```console
❯ make dummy-api

❯ ./bin/dummy-api -failureRate 0 &

❯ curl -sG http://localhost:8080/time -d tz=America/New_York | jq -r '"\(.zone) \(.unix)"'
EDT 1588641463

❯ curl -sG http://localhost:8080/time -d tz=Europe/Vilnius | jq -r '"\(.zone) \(.unix)"'
EEST 1588641513

❯ curl -sG http://localhost:8080/time -d tz=Chile/EasterIsland | jq .
{
  "now": "2020-05-04T19:22:29.6771843-06:00",
  "zone": "-06",
  "offset": -21600,
  "utc": "2020-05-05T01:22:29.6771843Z",
  "unix": 1588641749,
  "unixNano": 1588641749677184300
}

❯ kill -TERM $!
```

For more documentation about valid values for `tz`, please see the [godoc for
the `api`
package](https://pkg.go.dev/github.com/andreykaipov/http-flooder/dummy-api/api?tab=doc).

## Docker

The following is a manual walkthrough of the contents of
[`./test.sh`](./test.sh) as it's essentially a full integration test.

Build our images:

```console
❯ make images
```

Create a Docker network for both our applications:

```console
❯ docker network create wow
```

Start up the dummy API, serving responses randomly with a rather large failure
rate of 27%, and a delay interval of anywhere from 0 to 50 milliseconds:

```console
❯ docker run --rm --detach --network=wow --name=api dummy-api -failureRate 0.27 -delayInterval 0,50
```

Once we're done testing, don't forget to stop this container via `docker stop
api`. For now, let's continue and run the flooder against our API for 100
seconds at 100 requests per second:

```console
❯ docker run --rm --network=wow flooder -endpoint http://api:8080/time -duration 100 -requestsPerSecond 100 2>/dev/null
Starting flood. :-)
Running for 100 second(s), initiating 100 request(s) per second. Total requests
send to server will be 10000.
Sending batch 1
Sending batch 2
...
Sending batch 99
Sending batch 100

Total Requests: 10000
     Successes: 7311
      Failures: 2689
  Success Rate: 73.11000%
  Failure Rate: 26.89000%
  Average TTFB: 51.030573ms
  Average TTLB: 51.820745ms
         Delta: 790.172µs
```

Cool! Looks like the reported failure rate closely matches to what our dummy API
was invoked with. However, changes in the delay interval and how it affects the
TTFB and TLLB is a bit more subtle since concurrent requests also slow server
response times too.

For example, if we restart the dummy API with a delay interval of 0ms, and rerun
the above flood, we'll get an average TTxB of about 4xms, not quite 25ms (the
average of 0ms and 50ms) less than the 51ms TTFxB above. In fact, the delta
between the TTLB and TTFB seems to increase if there's no delay interval! Since
the delay is just implemented as a sleep, the most likely explanation is the OS
is dedicating the CPU cycles from the sleeping goroutines to accepting more
concurrent connections instead. This delay increases TTFB, but decreases the
relative TTLB. Very cool!

## TODO

- add support for retrying failed requests.

- add more information about the HTTP request (e.g. cover all of the events
  supported by [httptrace.ClientTrace](https://golang.org/pkg/net/http/httptrace/#ClientTrace)).

- abstract the `flooder`'s `request-per-second` parameter into a `batch-size`
  and `interval` paremeter, so that users can have more control over the flood
  rate.

- add a pipeline for building and pushing the images up to a Docker registry.
