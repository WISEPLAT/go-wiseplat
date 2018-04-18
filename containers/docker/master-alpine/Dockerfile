FROM alpine:3.5

RUN \
  apk add --update go git make gcc musl-dev linux-headers ca-certificates && \
  git clone --depth 1 --branch release/1.7 https://github.com/wiseplat/go-wiseplat && \
  (cd go-wiseplat && make gwsh) && \
  cp go-wiseplat/build/bin/gwsh /gwsh && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /go-wiseplat && rm -rf /var/cache/apk/*

EXPOSE 8747
EXPOSE 30373

ENTRYPOINT ["/gwsh"]
