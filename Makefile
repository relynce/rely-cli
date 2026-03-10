.PHONY: build install clean

GIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags "-X main.version=$(GIT_VERSION) -X main.gitHash=$(GIT_HASH)" -o rely ./cmd/rely

install: build
	sudo cp rely /usr/local/bin/rely
	@echo "Installed rely CLI to /usr/local/bin/rely"

clean:
	rm -f rely
