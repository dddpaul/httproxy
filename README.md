httproxy
=========

Simple HTTP reverse proxy with load balancing and logging capabilities written in Go.

## Features

- **Load Balancing**: Random load balancing across multiple upstream servers
- **Basic Authentication**: Supports upstream servers with basic auth (embed credentials in URL)
- **Request Logging**: Configurable logging with request details and request dumping
- **Error Handling**: Customizable error responses with custom status codes and body content
- **Redirect Following**: Optional internal following of 3xx redirects
- **Timeouts**: Configurable request timeouts
- **Request Dumping**: Full HTTP request dumping for debugging

## Installation

### From Source
```
go get -u github.com/dddpaul/httproxy
```

### Docker
```
docker pull dddpaul/httproxy
```

## Usage

```
httproxy [OPTIONS]
```

### Command Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `-port string` | `:8080` | Port to listen (prepended by colon), i.e. :8080 |
| `-url value` | *required* | List of URL to proxy to, i.e. http://localhost:8081 (can be specified multiple times) |
| `-timeout int` | `0` | Proxy request timeout (ms), 0 means no timeout |
| `-error-response-code int` | `502` | Override HTTP response code on proxy error |
| `-error-response-body string` | `""` | Body content on proxy error |
| `-follow` | `false` | Follow 3xx redirects internally |
| `-verbose` | `false` | Print request details |
| `-dump` | `false` | Dump request body |
| `-prefix string` | `httproxy` | Logging prefix |

### Examples

#### Basic Proxy
```bash
httproxy -url http://localhost:8081
```

#### Load Balancing Multiple Upstreams
```bash
httproxy -url http://server1:8081 -url http://server2:8081 -url http://server3:8081
```

#### Custom Port and Timeout
```bash
httproxy -port :9090 -url http://localhost:8081 -timeout 5000
```

#### With Authentication and Verbose Logging
```bash
httproxy -url http://user:pass@localhost:8081 -verbose
```

#### Custom Error Handling
```bash
httproxy -url http://localhost:8081 -error-response-code 503 -error-response-body "Service Temporarily Unavailable"
```

#### Debug Mode with Request Dumping
```bash
httproxy -url http://localhost:8081 -dump -verbose
```

### Docker Usage

#### Basic Usage
```bash
docker run -p 8080:8080 dddpaul/httproxy -url http://your-backend:8081
```

#### With Custom Configuration
```bash
docker run -p 9090:9090 dddpaul/httproxy \
  -port :9090 \
  -url http://backend1:8081 \
  -url http://backend2:8081 \
  -timeout 10000 \
  -verbose
```

## Load Balancing

The proxy uses simple random load balancing when multiple upstream URLs are provided. Each request is randomly distributed among the available upstream servers.

## Authentication

Basic authentication is supported by embedding credentials in the upstream URL:
```bash
httproxy -url http://username:password@upstream-server:8080
```

## Logging

The proxy provides two levels of logging:
- **Verbose mode** (`-verbose`): Logs request details including method, path, and response status
- **Dump mode** (`-dump`): Logs complete HTTP request including headers and body

## Building from Source

```bash
# Build for current platform
go build

# Build for Alpine Linux (used in Docker)
make build-alpine

# Build Docker image
make build
```

## License

This project is licensed under the Apache License 2.0 - see the LICENSE-2.0.txt file for details.