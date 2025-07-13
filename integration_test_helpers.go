package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// TestServer represents a test HTTP server with utility methods
type TestServer struct {
	*httptest.Server
	RequestCount int
	LastRequest  *http.Request
}

// NewTestServer creates a new test server with the given handler
func NewTestServer(handler http.HandlerFunc) *TestServer {
	ts := &TestServer{}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts.RequestCount++
		ts.LastRequest = r
		handler(w, r)
	}))
	return ts
}

// NewEchoServer creates a test server that echoes request information
func NewEchoServer() *TestServer {
	return NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Echo-Method", r.Method)
		w.Header().Set("Echo-Path", r.URL.Path)
		w.Header().Set("Echo-Query", r.URL.RawQuery)
		if auth := r.Header.Get("Authorization"); auth != "" {
			w.Header().Set("Echo-Auth", auth)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("echo response"))
	})
}

// NewSlowServer creates a test server that responds slowly
func NewSlowServer(delay time.Duration) *TestServer {
	return NewTestServer(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("slow response"))
	})
}

// NewErrorServer creates a test server that always returns an error
func NewErrorServer(statusCode int, message string) *TestServer {
	return NewTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(message))
	})
}

// NewRedirectServer creates a test server that redirects to another URL
func NewRedirectServer(redirectURL string) *TestServer {
	return NewTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", redirectURL)
		w.WriteHeader(http.StatusFound)
		_, _ = w.Write([]byte("redirect response"))
	})
}

// ProxyTestConfig holds configuration for proxy tests
type ProxyTestConfig struct {
	Timeout           int64
	ErrorResponseCode int
	ErrorResponseBody string
	FollowRedirects   bool
	Verbose           bool
	Dump              bool
}

// DefaultProxyTestConfig returns a default configuration for proxy tests
func DefaultProxyTestConfig() ProxyTestConfig {
	return ProxyTestConfig{
		Timeout:           0,
		ErrorResponseCode: http.StatusBadGateway,
		ErrorResponseBody: "",
		FollowRedirects:   false,
		Verbose:           false,
		Dump:              false,
	}
}

// SetupProxyTest configures global variables for proxy testing and returns a cleanup function
func SetupProxyTest(config ProxyTestConfig) func() {
	// store original values
	originalTimeout := timeout
	originalErrorCode := errorResponseCode
	originalErrorBody := errorResponseBody
	originalFollow := followRedirects
	originalVerbose := verbose
	originalDump := dump
	originalPrefix := log.Prefix()
	originalFlags := log.Flags()

	// set test values
	timeout = config.Timeout
	errorResponseCode = config.ErrorResponseCode
	errorResponseBody = config.ErrorResponseBody
	followRedirects = config.FollowRedirects
	verbose = config.Verbose
	dump = config.Dump

	// configure test logger
	log.SetPrefix("test: ")
	log.SetFlags(0)

	// return cleanup function
	return func() {
		timeout = originalTimeout
		errorResponseCode = originalErrorCode
		errorResponseBody = originalErrorBody
		followRedirects = originalFollow
		verbose = originalVerbose
		dump = originalDump
		log.SetPrefix(originalPrefix)
		log.SetFlags(originalFlags)
	}
}

// CreateTestURLs creates a slice of test URLs from server URLs
func CreateTestURLs(serverURLs ...string) []*url.URL {
	urls := make([]*url.URL, len(serverURLs))
	for i, serverURL := range serverURLs {
		u, err := url.Parse(serverURL)
		if err != nil {
			panic(err)
		}
		urls[i] = u
	}
	return urls
}

// AssertStatusCode checks if the response has the expected status code
func AssertStatusCode(t *testing.T, response *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if response.Code != expected {
		t.Errorf("Expected status code %d, got %d", expected, response.Code)
	}
}

// AssertResponseBody checks if the response body matches the expected content
func AssertResponseBody(t *testing.T, response *httptest.ResponseRecorder, expected string) {
	t.Helper()
	if response.Body.String() != expected {
		t.Errorf("Expected response body %q, got %q", expected, response.Body.String())
	}
}

// AssertHeader checks if a response header has the expected value
func AssertHeader(t *testing.T, response *httptest.ResponseRecorder, header, expected string) {
	t.Helper()
	actual := response.Header().Get(header)
	if actual != expected {
		t.Errorf("Expected header %s to be %q, got %q", header, expected, actual)
	}
}

// AssertRequestCount checks if a test server received the expected number of requests
func AssertRequestCount(t *testing.T, server *TestServer, expected int) {
	t.Helper()
	if server.RequestCount != expected {
		t.Errorf("Expected %d requests, got %d", expected, server.RequestCount)
	}
}

// CreateProxyRequest creates an HTTP request for testing proxy functionality
func CreateProxyRequest(method, path, body string) *http.Request {
	if body != "" {
		bodyReader := strings.NewReader(body)
		req := httptest.NewRequest(method, path, bodyReader)
		req.Header.Set("Content-Type", "text/plain")
		return req
	}
	return httptest.NewRequest(method, path, http.NoBody)
}

// RunProxyBehaviorTest is a helper function to test proxy behavior with various configurations
func RunProxyBehaviorTest(t *testing.T, name string, config ProxyTestConfig, backends []*TestServer, testFunc func(*testing.T, http.Handler)) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		cleanup := SetupProxyTest(config)
		defer cleanup()

		// create URLs from backend servers
		urls := make([]*url.URL, len(backends))
		for i, backend := range backends {
			u, _ := url.Parse(backend.URL)
			urls[i] = u
		}

		// create proxy
		proxy := newProxy(urls)

		// run the test
		testFunc(t, proxy)
	})
}

// IntegrationTestSuite provides a structure for running integration tests
type IntegrationTestSuite struct {
	t        *testing.T
	backends []*TestServer
	proxy    http.Handler
	cleanup  func()
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(t *testing.T, config ProxyTestConfig, backends []*TestServer) *IntegrationTestSuite {
	t.Helper()

	suite := &IntegrationTestSuite{
		t:        t,
		backends: backends,
	}

	suite.cleanup = SetupProxyTest(config)

	// create URLs from backend servers
	urls := make([]*url.URL, len(backends))
	for i, backend := range backends {
		u, _ := url.Parse(backend.URL)
		urls[i] = u
	}

	suite.proxy = newProxy(urls)
	return suite
}

// Close cleans up the test suite
func (suite *IntegrationTestSuite) Close() {
	suite.cleanup()
	for _, backend := range suite.backends {
		backend.Close()
	}
}

// SendRequest sends a request through the proxy and returns the response
func (suite *IntegrationTestSuite) SendRequest(method, path, body string) *httptest.ResponseRecorder {
	req := CreateProxyRequest(method, path, body)
	rr := httptest.NewRecorder()
	suite.proxy.ServeHTTP(rr, req)
	return rr
}

// AssertProxyResponse is a convenience method for asserting proxy responses
func (suite *IntegrationTestSuite) AssertProxyResponse(method, path string, expectedStatus int, expectedBody string) {
	suite.t.Helper()
	response := suite.SendRequest(method, path, "")
	AssertStatusCode(suite.t, response, expectedStatus)
	AssertResponseBody(suite.t, response, expectedBody)
}
