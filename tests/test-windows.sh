#!/bin/sh
DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$DIR/.."

git config --global user.email "githook@test.com" &&
    git config --global user.name "Githook Tests" &&
    git config --global init.defaultBranch master || exit 1

ROOT_DIR="C:/Program Files/githooks-tests"
export GITHOOKS_TEST_REPO="$ROOT_DIR/githooks"
export GITHOOKS_BIN_DIR="$GITHOOKS_TEST_REPO/githooks/bin"

if [ -d "$ROOT_DIR" ]; then
    echo "! You need to delete '$ROOT_DIR' first."
fi

mkdir -p "$GITHOOKS_TEST_REPO" || true
cp -rf "$REPO_DIR" "$GITHOOKS_TEST_REPO" || exit 1
rm -rf "$GITHOOKS_TEST_REPO/.git" || exit 1
# We use the bin folder.
sed -i -E 's/^bin//' "$GITHOOKS_TEST_REPO/githooks/.gitignore" || exit 1

echo "Make test gitrepo to clone from ..." &&
    cd "${GITHOOKS_TEST_REPO}" && git init &&
    git add . &&
    git commit -a -m "Before build" || exit 1

# Build binaries for v9.9.0-test.
#################################
cd "${GITHOOKS_TEST_REPO}/githooks" &&
    git tag "v9.9.0-test" &&
    ./scripts/clean.sh &&
    ./scripts/build.sh --build-flags "-tags debug,mock" &&
    ./bin/installer --version || exit 1
echo "Commit build v9.9.0-test to repo ..." &&
    cd "${GITHOOKS_TEST_REPO}" &&
    git add . &&
    git commit -a --allow-empty -m "Version 9.9.0-test" &&
    git tag -f "v9.9.0-test" || exit 1

# Build binaries for v9.9.1-test.
#################################
cd "${GITHOOKS_TEST_REPO}/githooks" &&
    git commit -a --allow-empty -m "Before build" &&
    git tag -f "v9.9.1-test" &&
    ./scripts/clean.sh &&
    ./scripts/build.sh --build-flags "-tags debug,mock" &&
    ./bin/installer --version || exit 1
echo "Commit build v9.9.1-test to repo ..." &&
    cd "${GITHOOKS_TEST_REPO}" &&
    git commit -a --allow-empty -m "Version 9.9.01test" &&
    git tag -f "v9.9.1-test" || exit 1

# Copy the tests somewhere different
cp -rf "$REPO_DIR/tests" "$ROOT_DIR/tests" || exit 1

if [ -n "${EXTRA_INSTALL_ARGS}" ]; then
    sed -i -E "s|(.*)/installer\"|\1/installer\" ${EXTRA_INSTALL_ARGS}|g" "$ROOT_DIR"/tests/step-*
fi

sh "$ROOT_DIR"/tests/exec-steps-go.sh --skip-docker-check
