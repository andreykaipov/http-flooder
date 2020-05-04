# build our go app
FROM golang:1.14-alpine as go
COPY . /app
WORKDIR /app
ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0
RUN go build -ldflags='-w -s' -o /main

# build our timezone info
# adapated from https://github.com/golang/go/tree/master/lib/time
FROM alpine as tzs
WORKDIR /app
ADD https://data.iana.org/time-zones/releases/tzcode2020a.tar.gz code.tgz
ADD https://data.iana.org/time-zones/releases/tzdata2020a.tar.gz data.tgz
RUN apk add -U make gcc musl-dev
RUN tar fx code.tgz
RUN tar fx data.tgz
RUN make CFLAGS=-DSTD_INSPIRED AWK=awk TZDIR=tzs posix_only

# copy the relevant ouputs from above
FROM scratch
COPY --from=go /main /dummy-api
COPY --from=tzs /app/tzs /tzs
ENV ZONEINFO=/tzs
ENTRYPOINT ["/dummy-api"]
