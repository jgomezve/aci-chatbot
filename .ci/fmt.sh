#!/bin/bash
#
# Code format check
GO_FILES=$(find . -name '*.go' | grep -v /vendor/ )

gofmt -d ${GO_FILES}

if [ "$(gofmt ${GO_FILES} | wc -l)" -gt 0 ]; then
     exit 1
fi