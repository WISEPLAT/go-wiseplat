# Build Gwsh in a stock Go builder container
FROM golang:1.9-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /go-wiseplat
RUN cd /go-wiseplat && make gwsh

# Pull Gwsh into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-wiseplat/build/bin/gwsh /usr/local/bin/

EXPOSE 8747 8748 30373 30373/udp 30374/udp
ENTRYPOINT ["gwsh"]
