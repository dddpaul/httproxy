package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/unrolled/logger"
)

const (
	defaultReadHeaderTimeout = 10 * time.Second
	defaultReadTimeout       = 30 * time.Second
	defaultWriteTimeout      = 30 * time.Second
	defaultIdleTimeout       = 60 * time.Second
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
	urls := make([]*url.URL, 0, len(*flags))
	for _, s := range *flags {
		u, err := url.Parse(s)
		if err != nil {
			panic(err)
		}
		urls = append(urls, u)
	}
	return urls
}

var (
	prefix            string
	verbose           bool
	dump              bool
	port              string
	urls              arrayFlags
	followRedirects   bool
	timeout           int64
	errorResponseCode int
	errorResponseBody string
	l                 *logger.Logger
)

func main() {
	flag.StringVar(&prefix, "prefix", "httproxy", "Logging prefix")
	flag.BoolVar(&verbose, "verbose", false, "Print request details")
	flag.BoolVar(&dump, "dump", false, "Dump request body")
	flag.StringVar(&port, "port", ":8080", "Port to listen (prepended by colon), i.e. :8080")
	flag.Var(&urls, "url", "List of URL to proxy to, i.e. http://localhost:8081")
	flag.BoolVar(&followRedirects, "follow", false, "Follow 3xx redirects internally")
	flag.Int64Var(&timeout, "timeout", 0, "Proxy request timeout (ms), 0 means no timeout")
	flag.IntVar(&errorResponseCode, "error-response-code", http.StatusBadGateway, "Override HTTP response code on proxy error")
	flag.StringVar(&errorResponseBody, "error-response-body", "", "Body content on proxy error")
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
	if dump {
		proxy = dumpMiddleware(proxy)
	}
	if verbose {
		proxy = l.Handler(proxy)
	}

	l.Printf("Proxy server is listening on port %s, upstreams = %s, timeout = %v ms, errorResponseCode = %v, followRedirects = %v, verbose = %v, dump = %v\n",
		port, urls, timeout, errorResponseCode, followRedirects, verbose, dump)
	server := &http.Server{
		Addr:              port,
		Handler:           proxy,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		ReadTimeout:       defaultReadTimeout,
		WriteTimeout:      defaultWriteTimeout,
		IdleTimeout:       defaultIdleTimeout,
	}
	l.Fatalln("ListenAndServe:", server.ListenAndServe())
}

func dumpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Printf("Failed to dump request: %v", err)
		} else {
			log.Println(string(dump))
		}
		next.ServeHTTP(w, r)
	})
}

func newProxy(urls []*url.URL) http.Handler {
	director := func(req *http.Request) {
		u := loadBalance(urls)
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		req.URL.Path = singleJoiningSlash(u.Path, req.URL.Path)
		req.Host = u.Host
		if u.User != nil {
			if pw, ok := u.User.Password(); ok {
				req.SetBasicAuth(u.User.Username(), pw)
			}
		}

		if timeout > 0 {
			ctx, cancel := context.WithTimeout(req.Context(), time.Duration(timeout)*time.Millisecond)
			defer cancel()
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
				return fmt.Errorf("failed to get response location: %w", err)
			}
		}

		r, err := http.Get(u.String())
		if err != nil {
			return fmt.Errorf("failed to follow redirect to %s: %w", u.String(), err)
		}

		cloneResponse(resp, r)
		return nil
	}

	errorHandler := func(rw http.ResponseWriter, _ *http.Request, err error) {
		l.Printf("Proxy error: %v\n", err)
		rw.WriteHeader(errorResponseCode)
		if errorResponseBody != "" {
			if _, err := rw.Write([]byte(errorResponseBody)); err != nil {
				l.Println(err)
			}
		}
	}

	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifier,
		ErrorHandler:   errorHandler,
	}
}

func loadBalance(targets []*url.URL) *url.URL {
	//nolint:gosec // Using weak random is acceptable for load balancing
	return targets[rand.IntN(len(targets))]
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
