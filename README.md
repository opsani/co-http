# co-http

A small http server for imitating simple workloads.

Build a container with the server:

```
docker build -t co-http .
```

Run a 2-tier app using the container:

```
docker network create -d bridge t
docker run -d --network=t --name back co-http
docker run -d --network=t -p 8080:8080 --name front co-http
```

Note the use of a user network (t) to allow containers to refer to each other by their name in network requests.

An example test request (the value of 'call' is a url-encoded url "http://back:8080/?busy=40"):

```
curl http://localhost:8080/?call=http%3A%2F%2Fback%3A8080%2F%3Fbusy%3D40\&busy=10
```

The URL request format:

`http://host:8080/?opt=`_value_[`&opt=`_value_...]

The following request parameters are supported and can be used in any combination:

- `busy=N` run a CPU-intensive operation N times (HMAC-SHA256 computation + forced garbage collector run)
- `call=URL` execute HTTP GET 'URL'
- `call=hostname` executes HTTP GET 'http://hostname:8080/' (value of `call` is assumed to be a hostname if it contains no slashes, otherwise it is treated as a URL, as above)
- `alloc=N` allocate N pages (4096 bytes each). NOTE this only performs an allocation, does not touch the data (this will typically NOT cause the process to add the given number of pages to its working set).
- `use=1` access every page allocated with alloc=N (can be used in the same URL or separately, the last allocated array will be accessed). This will trigger bringing in all pages in the allocated array into the process' working set.

When starting the server, one can also set a default query that will be used if the request URI has no query string in it:

```
docker run -d --network=t --name back co-http 'busy=20'
```

Now, when a request is made to `http://back:8080/`, it will behave as if the URI was `http://back:8080/?busy=20`.

The requests return plain text data (content-type: text/plain) with a short summary of every executed operation. If an error was encountered, the HTTP status will be 400 and the payload data will include a line prefixed with 'err: ...'.

By default, the server listens on all network interfaces on port 8080. This can be customized via `HTTP_ADDR` environment variable, e.g.:

```
docker run -d -p 8080:8081 -e HTTP_ADDR=:8081 --name front co-http
```

## Deploying 3-tier component stack with Kubernetes

See `k8s` sub-directory for sample app deployment with Optune servo for Kubernetes and using Apache Benchmark.

Replace fields next to `#CHANGEME` comment to set the app ID, account and auth token.