PROJECT=calendar-bot
ORGANISATION=three-men-in-a-boat
SOURCE=$(shell find . -name '*.go' | grep -v vendor/)
SOURCE_DIRS = cmd pkg

export GO111MODULE=on

.PHONY: vendor vetcheck fmtcheck clean build gotest

all: vendor vetcheck fmtcheck build gotest mod-clean

ver:
	@echo Building version: $(VERSION)

build: $(SOURCE)
	@mkdir -p build/bin
	go build -o build/bin/botbackend ./cmd/main.go

build-forkdetector-linux-amd64:
	@CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -o build/bin/linux-amd64/forkdetector -ldflags="-X main.version=$(VERSION)" ./cmd/forkdetector

build-forkdetector-linux-arm:
	@CGO_ENABLE=0 GOOS=linux GOARCH=arm go build -o build/bin/linux-arm/forkdetector -ldflags="-X main.version=$(VERSION)" ./cmd/forkdetector

gotest:
	go test -cover ./...

fmtcheck:
	@gofmt -l -s $(SOURCE_DIRS)

mod-clean:
	go mod tidy

clean:
	@rm -rf build
	go mod tidy

vendor:
	go mod vendor

vetcheck:
	go list ./... | grep -v bn254 | xargs go vet
	golangci-lint run --skip-dirs pkg/crypto/internal/groth16/bn256/utils/bn254