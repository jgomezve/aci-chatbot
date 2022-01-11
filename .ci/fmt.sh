#!/bin/bash
#
# Code format check
GO_FILES=$(find . -name '*.go' | grep -v /vendor/ )

unformatted=$(gofmt -l $GO_FILES)
[ -z "$unformatted" ] && exit 0

# Some files are not gofmt'd. Print message and fail.

echo >&2 "Go files must be formatted with gofmt. Please run:"
for fn in $unformatted; do
    echo >&2 "  gofmt -w $fn"
done

exit 1