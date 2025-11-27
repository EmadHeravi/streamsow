# ---------------------------------------------------------
# Streamzeug Makefile with version + git hash embedding
# ---------------------------------------------------------

SOURCES != find . -name '*.go'
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
DTAPI_INCLUDE = $(ROOT_DIR)/dektec/DTAPI/Include

VERSION := $(shell cat version/VERSION 2>/dev/null || echo "0.0.0")
GIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS = -ldflags "\
    -X github.com/EmadHeravi/streamsow/version.ProjectVersion=$(VERSION) \
    -X github.com/EmadHeravi/streamsow/version.GitVersion=$(GIT_HASH)"


all: binaries

# ---------------------------------------------------------
# Dektec / DTAPI
# ---------------------------------------------------------

dektecasi: dektec/asi.cpp dektec/asi.h
	CPATH=$(DTAPI_INCLUDE) g++ -c dektec/asi.cpp -o dektecasi.o

libdektec: dektecasi
	ar rcs ./libdektec.a dektecasi.o dektec/DTAPI/Lib/GCC5.1_CXX11_ABI1/DTAPI64.o

# ---------------------------------------------------------
# Streamzeug binary build
# ---------------------------------------------------------

binaries: streamzeug

streamzeug: bin/streamzeug

bin/streamzeug: $(SOURCES) go.mod go.sum libdektec
	LIBRARY_PATH=$(ROOT_DIR) go build $(LDFLAGS) -o bin/streamzeug ./cmd/streamzeug

# ---------------------------------------------------------
# Install target
# ---------------------------------------------------------

.PHONY: install
install: streamzeug
	install -o 0 -g 0 bin/streamzeug /usr/bin/streamzeug

# ---------------------------------------------------------
# Dev tools
# ---------------------------------------------------------

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run

# ---------------------------------------------------------
# Cleanup
# ---------------------------------------------------------

.PHONY: clean
clean:
	rm -rf bin libdektec.a dektecasi.*
