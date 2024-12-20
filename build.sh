#!/usr/bin/env bash
PROJ_DIR=$(cd ../;pwd)
export GOPATH=$GOPATH:$PROJ_DIR
ARCH=$(uname -m) \
    && echo "Architecture: $ARCH" \
    && GOOS=$(uname -s | tr '[:upper:]' '[:lower:]') \
    && echo "GOOS: $GOOS" \
    && if [ "$ARCH" = "x86_64" ]; then \
        GOARCH=amd64; \
      elif [ "$ARCH" = "aarch64" ]; then \
        GOARCH=arm64; \
      else \
        echo "Unsupported architecture: $ARCH"; \
        exit 1; \
      fi \
    && echo "GOARCH: $GOARCH" \
    && env


CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -o $1 main.go