FROM golang:1.13.6 AS builder
COPY http.go .
RUN go get -v -d .
RUN go build http.go

FROM ubuntu:16.04

COPY --from=builder /go/http /usr/local/bin/http

ENV HTTP_ADDR=:8080

ENTRYPOINT ["/usr/local/bin/http"]
