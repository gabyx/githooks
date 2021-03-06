#!/bin/sh

cat <<EOF | docker build --force-rm -t githooks:testsuite -
FROM golang:1.16-alpine
RUN apk add git curl git-lfs --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/main --allow-untrusted
RUN apk add bash jq

RUN curl -sSfL https://raw.githubusercontent.com/golangci/c/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v1.35.2

EOF

if ! docker run --rm -it \
    -v "$(pwd)":/githooks \
    -w /githooks githooks:testsuite \
    sh "tests/exec-testsuite.sh"; then

    echo "! Check rules had failures."
    exit 1
fi

exit 0
