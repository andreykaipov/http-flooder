# HTTP Flooder

Hello! This project contains:

1. A CLI to flood HTTP servers with requests (in the `flooder` directory)
1. A dummy webserver (in the `api` directory)

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
| requests-per-second | int    | number of GET requests per second to initiate against the endpoint | `1`     |
| timeout             | int    | timeout in milliseconds for the request to finish before failing   | `1000`  |
| verbose             | bool   | verbose logging while querying the servers                         | `false` |

After a run, the tool will output a summary about how well the server performed
against our flood of requests.

### Let's see

```console
❯ make flooder

❯ ./bin/flooder -endpoint http://google.com -requests-per-second 3 -duration 4 -report report.json
Starting flood. :-)
Running for 4 second(s), initiating 3 request(s) per second. Total requests send to server will be 12.
Sending batch 0
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

## API

### How does it work?

A dummy webserver, under the `api` directory, is what it sounds like. It'll spin
up a webserver serving one endpoint `/time` that will output time information
about whatever timezone (UTC by default) you pass it via the `tz` query
parameter.

TODO add reference to godoc here

In addition, we can customize how reliable our webserver behaves through the
following flags, namely `delay-interval` and `failure-rate`.

| flag           | type   | description                                                       | default   |
|----------------|--------|-------------------------------------------------------------------|-----------|
| bind           | string | the bind address and port for the server to listen on             | `":8080"` |
| delay-interval | string | add a random delay in milliseconds before processing a request    | `"0,100"` |
| failure-rate   | float  | percentage of requests to respond with 500s to, e.g. 0.13 for 13% | `0.1`     |

### Let's see

```console
❯ make api

❯ ./bin/api -failure-rate 0 &

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

❯ pkill api
```

## TODO

- add support for retrying failed requests.

- add more information about the HTTP request (e.g. cover all of the events
  supported by [httptrace.ClientTrace](https://golang.org/pkg/net/http/httptrace/#ClientTrace)).

- abstract the `flooder`'s `request-per-second` parameter into a `batch-size`
  and `interval` paremeter, so that users can have more control over the flood
  rate.
