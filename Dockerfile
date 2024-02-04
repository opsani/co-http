FROM golang:1.21-alpine AS builder
WORKDIR /src
COPY http.go otel.go go.mod go.sum startup.sh .
RUN chmod +x startup.sh
RUN go get -v -d .
RUN CGO_ENABLED=0 go build -o co-http

FROM alpine:3.19.1

COPY --from=builder /src/co-http /src/startup.sh /

ENV HTTP_ADDR=:8080

ENTRYPOINT ["/startup.sh"]
