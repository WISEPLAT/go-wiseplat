# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gwsh android ios gwsh-cross swarm evm all test clean
.PHONY: gwsh-linux gwsh-linux-386 gwsh-linux-amd64 gwsh-linux-mips64 gwsh-linux-mips64le
.PHONY: gwsh-linux-arm gwsh-linux-arm-5 gwsh-linux-arm-6 gwsh-linux-arm-7 gwsh-linux-arm64
.PHONY: gwsh-darwin gwsh-darwin-386 gwsh-darwin-amd64
.PHONY: gwsh-windows gwsh-windows-386 gwsh-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

gwsh:
	build/env.sh go run build/ci.go install ./cmd/gwsh
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gwsh\" to launch gwsh."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/gwsh.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Gwsh.framework\" to use the library."

test: all
#	build/env.sh go run build/ci.go test

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/jteeuwen/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go install ./cmd/abigen

# Cross Compilation Targets (xgo)

gwsh-cross: gwsh-linux gwsh-darwin gwsh-windows gwsh-android gwsh-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-*

gwsh-linux: gwsh-linux-386 gwsh-linux-amd64 gwsh-linux-arm gwsh-linux-mips64 gwsh-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-linux-*

gwsh-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/gwsh
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-linux-* | grep 386

gwsh-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/gwsh
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-linux-* | grep amd64

gwsh-linux-arm: gwsh-linux-arm-5 gwsh-linux-arm-6 gwsh-linux-arm-7 gwsh-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-linux-* | grep arm

gwsh-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/gwsh
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-linux-* | grep arm-5

gwsh-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/gwsh
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-linux-* | grep arm-6

gwsh-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/gwsh
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-linux-* | grep arm-7

gwsh-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/gwsh
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-linux-* | grep arm64

gwsh-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/gwsh
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-linux-* | grep mips

gwsh-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/gwsh
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-linux-* | grep mipsle

gwsh-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/gwsh
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-linux-* | grep mips64

gwsh-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/gwsh
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-linux-* | grep mips64le

gwsh-darwin: gwsh-darwin-386 gwsh-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-darwin-*

gwsh-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/gwsh
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-darwin-* | grep 386

gwsh-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gwsh
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-darwin-* | grep amd64

gwsh-windows: gwsh-windows-386 gwsh-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-windows-*

gwsh-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/gwsh
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-windows-* | grep 386

gwsh-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/gwsh
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gwsh-windows-* | grep amd64
