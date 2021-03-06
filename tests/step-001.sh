#!/bin/sh
# Test:
#   Run a simple install non-interactively and verify the hooks are in place

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

# run the default install
"$GH_TEST_BIN/cli" installer --non-interactive || exit 1

mkdir -p "$GH_TEST_TMP/test1" &&
    cd "$GH_TEST_TMP/test1" &&
    git init || exit 1

# verify that the pre-commit is installed
grep -q 'https://github.com/gabyx/githooks' .git/hooks/pre-commit
