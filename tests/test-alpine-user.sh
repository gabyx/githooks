#!/bin/sh

TEST_DIR=$(cd "$(dirname "$0")" && pwd)

cat <<EOF | docker build --force-rm -t githooks:alpine-user-base -
FROM golang:1.16-alpine
RUN apk add git git-lfs --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/main --allow-untrusted
RUN apk add bash jq
EOF

# shellcheck disable=SC2016,SC1004
export ADDITIONAL_PRE_INSTALL_STEPS='
RUN adduser -D -u 1099 test
RUN [ -d "$GH_TEST_GIT_CORE/templates/hooks" ] && \
    rm -rf "$GH_TEST_GIT_CORE/templates/hooks"
RUN mkdir -p "$GH_TEST_REPO" "$GH_TEST_GIT_CORE/templates/hooks" && \
    chown -R test:test "$GH_TEST_REPO" "$GH_TEST_GIT_CORE"
USER test
RUN mkdir -p /home/test/tmp
ENV GH_TEST_TMP=/home/test/tmp
'

exec sh "$TEST_DIR"/exec-tests.sh 'alpine-user' "$@"
