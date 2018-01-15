FROM ubuntu:16.04

COPY http /usr/local/bin/http

ENTRYPOINT ["/usr/local/bin/http"]
