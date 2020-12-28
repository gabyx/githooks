#!/bin/sh
DIR="$(cd "$(dirname "$0")" >/dev/null 2>&1 && pwd)"
GO_SRC="$DIR/.."

set -e
set -u

die() {
    echo "!! " "$@" >&2
    exit 1
}

BIN_DIR=""
BUILD_FLAGS=""

parseArgs() {
    prev_p=""
    for p in "$@"; do
        if [ "$p" = "--bin-dir" ]; then
            true
        elif [ "$prev_p" = "--bin-dir" ]; then
            BIN_DIR="$p"
        elif [ "$p" = "--build-flags" ]; then
            true
        elif [ "$prev_p" = "--build-flags" ]; then
            BUILD_FLAGS="$p"
        else
            echo "! Unknown argument \`$p\`" >&2
            return 1
        fi
        prev_p="$p"
    done
}

parseArgs "$@" || die "Parsing args failed."

cd "$GO_SRC"

export GOBIN="$GO_SRC/bin"
if [ -n "$BIN_DIR" ]; then
    if [ -d "$BIN_DIR" ]; then
        rm -rf "$BIN_DIR" || true
    fi
    export GOBIN="$BIN_DIR"
fi

if [ ! -d "$GO_SRC/vendor" ]; then
    echo "go vendor ..."
    go mod vendor
fi

echo "go install ..."
go generate -mod=vendor ./...
# shellcheck disable=SC2086
go install -mod=vendor \
    -tags debug $BUILD_FLAGS ./...