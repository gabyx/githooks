#!/bin/sh

TEST_DIR=$(cd "$(dirname "$0")" && pwd)

cat <<EOF | docker build --force-rm -t githooks:alpine-lfs-whitespace-base -
FROM golang:1.16-alpine
RUN apk add git git-lfs --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/main --allow-untrusted
RUN apk add bash jq

RUN mkdir -p "/root/whitespace folder"
ENV HOME="/root/whitespace folder"
EOF

# shellcheck disable=SC2016
export ADDITIONAL_INSTALL_STEPS='
# add a space in paths
RUN find "$GH_TESTS" -name "*.sh" -exec sed -i -E "s|GH_TEST_TMP(\})?/test([0-9.]+)|GH_TEST_TMP\1/test \2|g" {} \;
'

exec sh "$TEST_DIR"/exec-tests.sh 'alpine-lfs-whitespace' "$@"
