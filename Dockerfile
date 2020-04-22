FROM rust:1.40 as builder
WORKDIR /usr/src/myapp
COPY . .
RUN cargo install --path .

FROM debian:buster-slim

COPY --from=builder /usr/local/cargo/bin/co-http /usr/local/bin/co-http
ENTRYPOINT ["/usr/local/bin/co-http"]
