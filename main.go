package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var verbose bool
var port string
var upstreamURL string

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Print request details")
	flag.StringVar(&port, "port", ":8080", "Port to listen (prepended by colon), i.e. :8080")
	flag.StringVar(&upstreamURL, "url", "", "URL to proxy to, i.e. http://localhost:8081")
	flag.Parse()

	if len(upstreamURL) == 0 {
		panic("Upstream URL is mandatory")
	}

	upstream, err := url.Parse(upstreamURL)
	if err != nil {
		panic(err)
	}

	log.Printf("Proxy server is listening on port %s, upstream = %s, verbose = %v\n", port, upstreamURL, verbose)
	proxy := newProxy(upstream)
	http.ListenAndServe(port, proxy)
}

func newProxy(u *url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		if verbose {
			log.Printf("%s %s from [%s]\n", req.Method, req.RequestURI, req.RemoteAddr)
		}
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
	}
	return &httputil.ReverseProxy{
		Director: director,
	}
}
