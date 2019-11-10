httproxy
=========

Simple HTTP proxy with logging ability written in Go.

Install:

```
go get -u github.com/dddpaul/httproxy
```

Or grab Docker image:

```
docker pull dddpaul/httproxy
```

Usage:

```
httproxy [OPTIONS]
  -port string
      Port to listen (prepended by colon), i.e. :8080 (default ":8080")
  -url value
      List of URL to proxy to, i.e. http://localhost:8081
  -follow
      Follow 3xx redirects internally
  -verbose
      Print request details
```
