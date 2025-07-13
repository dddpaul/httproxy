package main

import (
	"net/url"
	"testing"
)

// BenchmarkLoadBalance tests the performance of the load balancing function
func BenchmarkLoadBalance(b *testing.B) {
	// create test URLs
	urls := make([]*url.URL, 10)
	for i := range 10 {
		u, _ := url.Parse("http://server" + string(rune('0'+i)) + ":8080")
		urls[i] = u
	}

	for b.Loop() {
		loadBalance(urls)
	}
}

// BenchmarkLoadBalanceSingle tests load balancing with a single URL
func BenchmarkLoadBalanceSingle(b *testing.B) {
	u, _ := url.Parse("http://localhost:8080")
	urls := []*url.URL{u}

	for b.Loop() {
		loadBalance(urls)
	}
}

// BenchmarkSingleJoiningSlash tests the URL path joining function
func BenchmarkSingleJoiningSlash(b *testing.B) {
	testCases := []struct {
		name string
		a    string
		b    string
	}{
		{"both_slashes", "/api/", "/users"},
		{"no_slashes", "api", "users"},
		{"a_slash", "/api/", "users"},
		{"b_slash", "api", "/users"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				singleJoiningSlash(tc.a, tc.b)
			}
		})
	}
}

// BenchmarkArrayFlagsString tests the string formatting of array flags
func BenchmarkArrayFlagsString(b *testing.B) {
	flags := arrayFlags{
		"http://server1:8080",
		"http://server2:8080",
		"http://server3:8080",
		"http://server4:8080",
		"http://server5:8080",
	}

	for b.Loop() {
		_ = flags.String()
	}
}

// BenchmarkArrayFlagsToURLs tests URL parsing performance
func BenchmarkArrayFlagsToURLs(b *testing.B) {
	flags := arrayFlags{
		"http://server1:8080",
		"http://server2:8080",
		"http://server3:8080",
		"http://server4:8080",
		"http://server5:8080",
	}

	for b.Loop() {
		// create a copy to avoid modifying the original
		testFlags := make(arrayFlags, len(flags))
		copy(testFlags, flags)
		_ = testFlags.toURLs()
	}
}

// BenchmarkLoadBalanceDistribution tests load balancing distribution over many calls
func BenchmarkLoadBalanceDistribution(b *testing.B) {
	// create test URLs
	urls := make([]*url.URL, 5)
	for i := range 5 {
		u, _ := url.Parse("http://server" + string(rune('0'+i)) + ":8080")
		urls[i] = u
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			loadBalance(urls)
		}
	})
}
