FROM golang:1.10.0 AS builder
COPY http.go .
RUN go get -v -d .
RUN go build http.go

FROM ubuntu:16.04

COPY --from=builder /go/http /usr/local/bin/http

ENTRYPOINT ["/usr/local/bin/http"]
