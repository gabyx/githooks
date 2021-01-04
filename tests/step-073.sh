#!/bin/sh
# Test:
#   Run the cli tool trying to list a not yet trusted repo

if ! "$GITHOOKS_BIN_DIR/installer" --stdin; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p /tmp/test073/.githooks/pre-commit &&
    echo 'echo "Hello"' >/tmp/test073/.githooks/pre-commit/testing &&
    touch /tmp/test073/.githooks/trust-all &&
    cd /tmp/test073 &&
    git init ||
    exit 1

if "$GITHOOKS_EXE_GIT_HOOKS" list pre-commit | grep -i "'trusted'"; then
    echo "! Unexpected list result"
    exit 1
fi
