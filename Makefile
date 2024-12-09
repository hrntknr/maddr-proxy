.PHONY: all build run test clean

GO_FILES   := $(shell find . -type f -name '*.go' -print)

GO_LDFLAGS  = -ldflags="-s -w"
TARGET      = bin/main

all: build run

build: $(TARGET)

run:
	./bin/main proxy

test:
	go test ./pkg/...

clean:
	rm -rf $(TARGET)

$(TARGET): $(GO_FILES)
	go build $(GO_LDFLAGS) -o $@ ./cmd/
