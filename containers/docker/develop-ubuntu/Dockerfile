FROM ubuntu:xenial

RUN \
  apt-get update && apt-get upgrade -q -y && \
  apt-get install -y --no-install-recommends golang git make gcc libc-dev ca-certificates && \
  git clone --depth 1 https://github.com/wiseplat/go-wiseplat && \
  (cd go-wiseplat && make gwsh) && \
  cp go-wiseplat/build/bin/gwsh /gwsh && \
  apt-get remove -y golang git make gcc libc-dev && apt autoremove -y && apt-get clean && \
  rm -rf /go-wiseplat

EXPOSE 8747
EXPOSE 30373

ENTRYPOINT ["/gwsh"]
