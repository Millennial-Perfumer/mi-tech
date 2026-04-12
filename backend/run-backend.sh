#!/bin/sh
# Path Shield Wrapper for Air
# This handles the project path spaces by using relative execution
CURRENT_DIR=$(dirname "$0")
cd "$CURRENT_DIR" || exit 1
exec ./tmp/main "$@"
