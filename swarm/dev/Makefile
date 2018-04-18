.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/wiseplat/go-wiseplat/cmd/swarm
	go build -o bin/gwsh     github.com/wiseplat/go-wiseplat/cmd/gwsh
	go build -o bin/bootnode github.com/wiseplat/go-wiseplat/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/wiseplat/go-wiseplat/swarm/...
