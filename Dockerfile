FROM golang:1.9.0 AS builder
WORKDIR /
COPY http.go .
RUN go build http.go

FROM ubuntu:16.04

COPY --from=builder /http /usr/local/bin/http

ENTRYPOINT ["/usr/local/bin/http"]
