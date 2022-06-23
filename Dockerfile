FROM golang:1.18-alpine AS builder
WORKDIR /src
COPY http.go go.mod go.sum .
RUN go get -v -d .
RUN CGO_ENABLED=0 go build http.go

FROM scratch

COPY --from=builder /src/http /

ENV HTTP_ADDR=:8080

ENTRYPOINT ["/http"]
