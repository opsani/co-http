# co-http

A small http server for imitating simple workloads.


_Note_: This is a very limited version of the co-http server that resides in the master branch. It supports only 
command-line parameters (no in-request query string) and only cpu load (`work=NNNNN`). This version is implemented
in Rust in order to test support more consistent per-request timing.

Build a container with the server:

```
docker build -t co-http .
```

Run a single-tier app using the container:

```
docker network create -d bridge t
docker run -d --network=t -p 8080:8080 --name front co-http work=10000
```

Note the use of a user network (t) to allow containers to refer to each other by their name in network requests.


The URL request format:

`http://host:8080/`

The following request parameters are supported and can be used in any combination:

- `work=N` run a CPU-intensive operation N times (HMAC-SHA256 computation)

When starting the server, one sets the amount of work that co-http will perform on each request:

```
docker run -d --network=t --name back co-http 'work=20000'
```

The requests return plain text data (content-type: text/plain) with a short summary of every executed operation. If an error was encountered, the HTTP status will be 400 and the payload data will include a line prefixed with 'err: ...'.

By default, the server listens on all network interfaces on port 8080.

