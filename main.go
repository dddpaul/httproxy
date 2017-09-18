package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type arrayFlags []string

func (flags *arrayFlags) String() string {
	return strings.Join(*flags, ", ")
}

func (flags *arrayFlags) Set(value string) error {
	*flags = append(*flags, value)
	return nil
}

func (flags *arrayFlags) toURLs() []*url.URL {
	var urls []*url.URL
	for _, s := range *flags {
		u, err := url.Parse(s)
		if err != nil {
			panic(err)
		}
		urls = append(urls, u)
	}
	return urls
}

var verbose bool
var port string
var urls arrayFlags

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Print request details")
	flag.StringVar(&port, "port", ":8080", "Port to listen (prepended by colon), i.e. :8080")
	flag.Var(&urls, "url", "List of URL to proxy to, i.e. http://localhost:8081")
	flag.Parse()

	if len(urls) == 0 {
		panic("At least on URL has to be specified")
	}

	log.Printf("Proxy server is listening on port %s, upstreams = %s, verbose = %v\n", port, urls, verbose)
	proxy := newProxy(urls.toURLs())
	http.ListenAndServe(port, proxy)
}

func newProxy(urls []*url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		u := loadBalance(urls)
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		req.URL.Path = singleJoiningSlash(u.Path, req.URL.Path)
		if verbose {
			log.Printf("%s %s from [%s] passed to %v\n", req.Method, req.RequestURI, req.RemoteAddr, req.URL)
		}
	}

	return &httputil.ReverseProxy{
		Director: director,
	}
}

func loadBalance(targets []*url.URL) *url.URL {
	return targets[rand.Int()%len(targets)]
}

// Taken from net/http/httputil/reverseproxy.go
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
