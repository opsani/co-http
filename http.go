package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var m []byte
var dflt_qry string

type ApiHandler struct{}

func (ApiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var v string
	var a, u bool
	var data string
	var http_status int

	qry := r.URL.RawQuery
	if qry == "" {
		qry = dflt_qry
	}
	vals, _ := url.ParseQuery(qry) // map[string] []string ; NB: use vals.Get() to get the 1st value

	http_status = 200

	if v = vals.Get("busy"); v != "" {
		t := time.Now()
		c, _ := strconv.ParseUint(v, 10, 64)
		k := make([]byte, 32)
		mac := hmac.New(sha256.New, k)
		for i := 0; i < int(c); i++ {
			mac.Write(k)
			mac.Write(k)
			runtime.GC()
		}
		data += fmt.Sprintf("busy for %d us\n", time.Now().Sub(t)/1000)
	}

	if v = vals.Get("call"); v != "" {
		if !strings.ContainsRune(v, '/') {
			v = fmt.Sprintf("http://%s:8080/", v)
		}
		rsp, err := http.Get(v)
		if err != nil {
			data = "err: " + err.Error() + "\n"
			http_status = 400
		} else {
			var d []byte
			d, err = io.ReadAll(rsp.Body) // TODO: err
			data += "call: " + string(d)
		}
	}

	if v = vals.Get("alloc"); v != "" {
		a = true
	}
	if vals.Get("use") != "" {
		u = true
	}

	if a {
		var sz uint64
		sz, _ = strconv.ParseUint(v, 10, 64)
		//@err
		m = nil
		runtime.GC()
		m = make([]byte, sz*4096)
		data += fmt.Sprintf("allocated memory %d bytes (%d pages)\n", len(m), len(m)/4096)
	}
	if u {
		var i int
		for i = 0; i < len(m); i += 4096 {
			m[i] = 1
		}
		data += fmt.Sprintf("accessed %d bytes (%d pages)\n", len(m), len(m)/4096)
	}

	w.Header().Add("Content-length", strconv.Itoa(len(data)))
	w.Header().Add("Content-type", "text/plain")
	w.WriteHeader(http_status)
	w.Write([]byte(data))
}

func main() {
	// --- Parse command line

	// server address
	serverAddr := ":8080"
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		serverAddr = addr
	}

	// default query from command line
	if len(os.Args) > 1 {
		dflt_qry = os.Args[1]
	}

	// --- Prepare for serving

	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		die("failed to set up OpenTelemetry:", err)
		return // keep linters happy
	}

	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// setup a Prometheus metric
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)
	prometheus.MustRegister(counter)
	apiHandlerFn := promhttp.InstrumentHandlerCounter(counter, ApiHandler{})

	// Start HTTP server.
	runtime.GC()
	srv := &http.Server{
		Addr:         serverAddr,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      coHTTPHandler(apiHandlerFn, promhttp.Handler()),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		die("Error starting HTTP server:", err)
		return
	case <-ctx.Done():
		// CTRL+C received.
		slog.Warn("Shutting down...")
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	if err != nil && errors.Is(err, http.ErrServerClosed) {
		die("Error shutting down HTTP server:", err)
	}
}

func coHTTPHandler(mainHandlerFn http.HandlerFunc, promHandler http.Handler) http.Handler {
	mux := http.NewServeMux()

	// handle is a replacement for mux.Handle
	// that enriches the handler's HTTP instrumentation with the pattern as the http.route tag
	handle := func(pattern string, h http.Handler) {
		handler := otelhttp.WithRouteTag(pattern, h)
		mux.Handle(pattern, handler)
	}

	// Register handlers
	handle("/", http.HandlerFunc(mainHandlerFn))
	handle("/metrics", promHandler)

	// Add HTTP instrumentation for the whole server
	handler := otelhttp.NewHandler(mux, "/")
	return handler
}

func die(fmt string, args ...interface{}) {
	slog.Error(fmt, args...)
	os.Exit(1)
}
