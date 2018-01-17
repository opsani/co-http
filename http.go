package main

import (
	"net/http"
//	"crypto/tls"
//	"crypto/x509"
	"strconv"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"io/ioutil"
	"os"

	"time"
	"crypto/hmac"
	"crypto/sha256"
)


	const cfile = "/home/lka/dg/skopos/cfile.pem"

var m []byte
var dflt_qry string

func reply(w http.ResponseWriter, r *http.Request) {
	var v string
	var a, u bool
	var data string
	var http_status int

	qry := r.URL.RawQuery
	if qry == "" { qry = dflt_qry }
	vals,_ := url.ParseQuery(qry) // map[string] []string ; NB: use vals.Get() to get the 1st value

	http_status = 200

	if v=vals.Get("busy") ; v != "" {
		t := time.Now()
		c, _ := strconv.ParseUint(v, 10, 64)
		k := make([]byte, 32)
		mac := hmac.New(sha256.New, k)
		for i :=0 ; i<int(c) ; i++ {
			mac.Write(k)
			mac.Write(k)
			runtime.GC()
		}
		data += fmt.Sprintf("busy for %d us\n", time.Now().Sub(t)/1000)
	}

	if v=vals.Get("call") ; v != "" {
		if !strings.ContainsRune(v, '/') {
			v = fmt.Sprintf("http://%s:8080/",v)
		}
		rsp, err := http.Get(v)
		if err != nil {
			data = "err: " + err.Error() +"\n"
			http_status = 400
		} else {
			var d []byte
			d,err = ioutil.ReadAll(rsp.Body) // TODO: err
			data += "call: " + string(d)
		}
	}

	if v=vals.Get("alloc") ; v != "" {
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
		m = make([]byte,sz*4096)
		data += fmt.Sprintf("allocated memory %d bytes (%d pages)\n", len(m), len(m)/4096)
	}
	if u {
		var i int
		for i=0 ; i<len(m) ; i+=4096 {
			m[i]=1
		}
		data += fmt.Sprintf("accessed %d bytes (%d pages)\n", len(m), len(m)/4096)
	}

	w.Header().Add("Content-length", strconv.Itoa(len(data)))
	w.Header().Add("Content-type", "text/plain")
	w.WriteHeader(http_status)
	w.Write([]byte(data))
}

func main() {
	var s http.Server
	var err error
	s.Addr = ":8080"

	// default query from command line
	if len(os.Args)>1 { dflt_qry = os.Args[1] }

	/*
	// We could call ListenAndServeTLS without initializing the TLS config, but we 
	// set it up explicitly here anyway, to allow adding things like client cert verify, etc. later
	var tlscfg tls.Config
	tlscfg.Certificates = make([]tls.Certificate, 1)
	tlscfg.Certificates[0], err = tls.LoadX509KeyPair(cfile,cfile)
	// @@ err
	// pre-init the leaf cert, too (so it isn't parsed every time when serving requests)
	tlscfg.Certificates[0].Leaf, err = x509.ParseCertificate(tlscfg.Certificates[0].Certificate[0])
	// @@ err
	s.TLSConfig = &tlscfg
	*/

	// Handler - default
	http.HandleFunc("/", reply)
	// err = s.ListenAndServeTLS("","")
	runtime.GC()
	err = s.ListenAndServe()
	fmt.Println(err)
}
