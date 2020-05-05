FROM golang:1.14-alpine as go
COPY . /app/
WORKDIR /app
ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0
RUN go build -ldflags='-w -s' -o /main

FROM alpine as certs
RUN apk add -U ca-certificates

FROM scratch
COPY --from=go /main /flooder
COPY --from=certs /etc/ssl/certs /etc/ssl/certs
ENTRYPOINT ["/flooder"]
