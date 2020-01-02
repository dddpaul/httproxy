FROM golang:1.13.5 as builder
WORKDIR /go/src/github.com/dddpaul/httproxy
ADD . ./
RUN make build-alpine

FROM alpine:latest
WORKDIR /app
COPY --from=builder /go/src/github.com/dddpaul/httproxy/bin/httproxy .
EXPOSE 8080

ENTRYPOINT ["./httproxy"]
CMD ["-port", ":8080"]
