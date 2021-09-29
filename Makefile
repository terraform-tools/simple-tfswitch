EXE  := simple-tfswitch
PKG  := github.com/terraform-tools/simple-tfswitch
VER := $(shell cat version)
PATH := build:$(PATH)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

$(EXE): go.mod *.go pkg/*.go
	go build -v -ldflags "-X main.version=$(VER)" -o $@ $(PKG)

.PHONY: release
release: $(EXE) darwin linux

.PHONY: darwin linux 
darwin linux:
	GOOS=$@ go build -ldflags "-X main.version=$(VER)" -o $(EXE)-$(VER)-$@-$(GOARCH) $(PKG)

.PHONY: clean
clean:
	rm -f $(EXE) $(EXE)-*-*-*

.PHONY: test
test:
	go test -v ./...
