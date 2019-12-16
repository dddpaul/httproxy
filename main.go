package main

import (
	"context"
	"flag"
	"github.com/unrolled/logger"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
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

var prefix string
var verbose bool
var port string
var urls arrayFlags
var followRedirects bool
var timeout int64
var errorResponseCode int
var l *logger.Logger

func main() {
	flag.StringVar(&prefix, "prefix", "httproxy", "Logging prefix")
	flag.BoolVar(&verbose, "verbose", false, "Print request details")
	flag.StringVar(&port, "port", ":8080", "Port to listen (prepended by colon), i.e. :8080")
	flag.Var(&urls, "url", "List of URL to proxy to, i.e. http://localhost:8081")
	flag.BoolVar(&followRedirects, "follow", false, "Follow 3xx redirects internally")
	flag.Int64Var(&timeout, "timeout", 0, "Proxy request timeout (ms), 0 means no timeout")
	flag.IntVar(&errorResponseCode, "error-response-code", http.StatusBadGateway, "Override HTTP response code on proxy error")
	flag.Parse()

	if len(urls) == 0 {
		panic("At least on URL has to be specified")
	}

	l = logger.New(logger.Options{
		Prefix:               prefix,
		RemoteAddressHeaders: []string{"X-Forwarded-For"},
		OutputFlags:          log.LstdFlags,
	})

	proxy := newProxy(urls.toURLs())
	if verbose {
		proxy = l.Handler(proxy)
	}

	l.Printf("Proxy server is listening on port %s, upstreams = %s, timeout = %v ms, errorResponseCode = %v, followRedirects = %v, verbose = %v\n",
		port, urls, timeout, errorResponseCode, followRedirects, verbose)
	l.Fatalln("ListenAndServe:", http.ListenAndServe(port, proxy))
}

func newProxy(urls []*url.URL) http.Handler {
	director := func(req *http.Request) {
		u := loadBalance(urls)
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		req.URL.Path = singleJoiningSlash(u.Path, req.URL.Path)
		req.Host = u.Host

		if timeout > 0 {
			ctx, _ := context.WithTimeout(req.Context(), time.Duration(timeout)*time.Millisecond)
			req2 := req.WithContext(ctx)
			*req = *req2
		}
	}

	modifier := func(resp *http.Response) error {
		if !followRedirects {
			return nil
		}

		u, err := resp.Location()
		if err != nil {
			switch err {
			case http.ErrNoLocation:
				return nil
			default:
				return err
			}
		}

		r, err := http.Get(u.String())
		if err != nil {
			return err
		}

		cloneResponse(resp, r)
		return nil
	}

	errorHandler := func(rw http.ResponseWriter, req *http.Request, err error) {
		l.Printf("Proxy error: %v\n", err)
		rw.WriteHeader(errorResponseCode)
	}

	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifier,
		ErrorHandler:   errorHandler,
	}
}

func loadBalance(targets []*url.URL) *url.URL {
	return targets[rand.Int()%len(targets)]
}

func cloneResponse(to, from *http.Response) {
	to.Status = from.Status
	to.StatusCode = from.StatusCode
	to.Body = from.Body
	to.ContentLength = from.ContentLength
	headers := []string{"Content-Length", "Content-Encoding", "Content-Type"}
	for _, h := range headers {
		replaceHeader(to, from, h)
	}
	to.Header.Del("Location")
}

func replaceHeader(to, from *http.Response, header string) {
	if from.Header.Get(header) != "" {
		to.Header.Set(header, from.Header.Get(header))
	} else {
		to.Header.Del(header)
	}
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
