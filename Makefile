.PHONY: all build run test clean

GO_FILES   := $(shell find . -type f -name '*.go' -print)

GO_LDFLAGS  = -ldflags="-s -w"
TARGET      = bin/main

all: build run

build: $(TARGET)

run:
	./bin/main

test:
	go test ./pkg/...

clean:
	rm -rf $(TARGET)

$(TARGET): $(GO_FILES)
	go build $(GO_LDFLAGS) -tags=viper_bind_struct -o $@ ./cmd/
