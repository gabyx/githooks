#!/bin/sh
TEST_DIR=$(cd "$(dirname "$0")" && pwd)

cat <<EOF | docker build --force-rm -t githooks:alpine-coverage-base -
FROM golang:1.16-alpine
RUN apk add git git-lfs --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/main --allow-untrusted
RUN apk add bash jq
RUN go get github.com/wadey/gocovmerge
RUN go get github.com/mattn/goveralls

ENV GH_COVERAGE_DIR="/cover"

# Coveralls env. vars for upload.
ENV COVERALLS_TOKEN="$COVERALLS_TOKEN"
EOF

IMAGE_TYPE="alpine-coverage"

cleanup() {
    docker rmi "githooks:$IMAGE_TYPE-base"
    docker rmi "githooks:$IMAGE_TYPE"
}

trap cleanup EXIT

if echo "$IMAGE_TYPE" | grep -q "\-user"; then
    OS_USER="test"
else
    OS_USER="root"
fi

cat <<EOF | docker build --force-rm -t githooks:$IMAGE_TYPE -f - .
FROM githooks:$IMAGE_TYPE-base

ENV GH_TESTS="/var/lib/githooks-tests"
ENV GH_TEST_TMP="/tmp"
ENV GH_TEST_REPO="/var/lib/githooks"
ENV GH_TEST_BIN="/var/lib/githooks/githooks/bin"
ENV GH_TEST_GIT_CORE="/usr/share/git-core"

${ADDITIONAL_PRE_INSTALL_STEPS:-}

# Add sources.
COPY --chown=$OS_USER:$OS_USER githooks "\$GH_TEST_REPO/githooks"
RUN sed -i -E 's/^bin//' "\$GH_TEST_REPO/githooks/.gitignore" # We use the bin folder
ADD .githooks/README.md \$GH_TEST_REPO/.githooks/README.md
ADD examples "\$GH_TEST_REPO/examples"
ADD tests "\$GH_TESTS"

RUN git config --global user.email "githook@test.com" && \\
    git config --global user.name "Githook Tests" && \\
    git config --global init.defaultBranch main && \\
    git config --global core.autocrlf false

RUN echo "Make test gitrepo to clone from ..." && \\
    cd \$GH_TEST_REPO && git init  >/dev/null 2>&1  && \\
    git add . >/dev/null 2>&1  && \\
    git commit -a -m "Before build" >/dev/null 2>&1

# Replace run-wrapper with coverage run-wrapper
RUN cd \$GH_TEST_REPO && \\
    cp githooks/build/embedded/run-wrapper-coverage.sh githooks/build/embedded/run-wrapper.sh

# Build binaries for v9.9.0.
#################################
RUN cd \$GH_TEST_REPO/githooks && \\
    git tag "v9.9.0" >/dev/null 2>&1 && \\
     ./scripts/clean.sh && \\
    ./scripts/build.sh --coverage --build-flags "-tags debug,mock,coverage" && \\
    ./bin/cli githooksCoverage --version
RUN echo "Commit build v9.9.0 to repo ..." && \\
    cd \$GH_TEST_REPO && \\
    git add . >/dev/null 2>&1 && \\
    git commit -a --allow-empty -m "Version 9.9.0" >/dev/null 2>&1 && \\
    git tag -f "v9.9.0"

# Build binaries for v9.9.1.
#################################
RUN cd \$GH_TEST_REPO/githooks && \\
    git commit -a --allow-empty -m "Before build" >/dev/null 2>&1 && \\
    git tag -f "v9.9.1" && \\
    ./scripts/clean.sh && \\
    ./scripts/build.sh --coverage --build-flags "-tags debug,mock,coverage" && \\
    ./bin/cli githooksCoverage --version
RUN echo "Commit build v9.9.1 to repo (no-skip)..." && \\
    cd "\$GH_TEST_REPO" && \\
    git commit -a --allow-empty -m "Version 9.9.1" -m "Update-NoSkip: true" >/dev/null 2>&1 && \\
    git tag -f "v9.9.1"

# Commit for to v9.9.2 (not used for update).
#################################
RUN echo "Commit build v9.9.2 to repo ..." && \\
    cd "\$GH_TEST_REPO" && \\
    git commit -a --allow-empty -m "Version 9.9.2" >/dev/null 2>&1 && \\
    git tag -f "v9.9.2"
RUN if [ -n "\$EXTRA_INSTALL_ARGS" ]; then \\
        sed -i -E 's|(.*)/cli" installer|\1/cli" installer \$EXTRA_INSTALL_ARGS|g' "\$GH_TESTS"/step-* ; \\
    fi

# Always don't delete LFS Hooks (for testing, default is unset, but cumbersome for tests)
RUN git config --global githooks.deleteDetectedLFSHooks "n"

# Replace some statements which rely on proper CLI output
# The built instrumented executable output test&coverage shit...
RUN sed -i -E 's@cli" shared location(.*)\)@cli" shared location\1 | grep "^/")@g' \\
    "\$GH_TESTS"/step-*

# Replace all runnner/cli/dialog/'git hooks' invocations.
# Foward over 'coverage/forwarder'.
RUN sed -i -E 's@"(.GH_INSTALL_BIN_DIR|.GH_TEST_BIN)/cli"@"\$GH_TEST_REPO/githooks/coverage/forwarder" "\1/cli"@g' \\
    "\$GH_TESTS/exec-steps.sh" \\
    "\$GH_TESTS"/step-*
RUN sed -i -E 's@"(.GH_INSTALL_BIN_DIR|.GH_TEST_BIN)/runner"@"\$GH_TEST_REPO/githooks/coverage/forwarder" "\1/runner"@g' \\
    "\$GH_TESTS"/step-*
RUN sed -i -E 's@"(.GH_INSTALL_BIN_DIR|.GH_TEST_BIN)/dialog"@"\$GH_TEST_REPO/githooks/coverage/forwarder" "\1/dialog"@g' \\
    "\$GH_TESTS"/step-*
RUN sed -i -E 's@".DIALOG"@"\$GH_TEST_REPO/githooks/coverage/forwarder" "\1"@g' \\
    "\$GH_TESTS"/step-*
RUN sed -i -E 's@git hooks@"\$GH_TEST_REPO/githooks/coverage/forwarder" "\$GH_INSTALL_BIN_DIR/cli"@g' \\
    "\$GH_TESTS"/step-*


${ADDITIONAL_INSTALL_STEPS:-}

RUN echo "Git version: \$(git --version)"
WORKDIR \$GH_TESTS
EOF

# Clean all coverage data
if [ -d "$TEST_DIR/cover" ]; then
    rm -rf "$TEST_DIR/cover"/*
fi

# Run the normal tests to add to the coverage
# inside the current repo
docker run --rm \
    -a stdout -a stderr \
    -v "$TEST_DIR/cover":/cover \
    -v "$TEST_DIR/..":/githooks \
    -w /githooks/tests \
    "githooks:$IMAGE_TYPE-base" \
    sh ./exec-testsuite.sh ||
    exit $?

# Run the integration tests
docker run --rm \
    -a stdout -a stderr \
    -v "$TEST_DIR/cover":/cover \
    "githooks:$IMAGE_TYPE" \
    ./exec-steps.sh "$@" || exit $?

# Upload the produced coverage
# inside the current repo
docker run --rm \
    -a stdout -a stderr \
    -v "$TEST_DIR/cover":/cover \
    -v "$TEST_DIR/..":/githooks \
    -w /githooks \
    "githooks:$IMAGE_TYPE-base" \
    ./tests/upload-coverage.sh "$@" || exit $?
