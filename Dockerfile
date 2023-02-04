FROM golang:1.19.5 as builder
WORKDIR /go/src/github.com/dddpaul/httproxy
ADD . ./
RUN make build-alpine

FROM alpine:3.16.3
LABEL maintainer="Pavel Derendyaev <dddpaul@gmail.com>"
RUN addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app
COPY --from=builder /go/src/github.com/dddpaul/httproxy/bin/httproxy .
EXPOSE 8080

CMD ["./httproxy", "-port", ":8080"]
