#!/bin/sh

GOIMPORTS_FILES=$(gofmt -l .)
if [ -n "${GOIMPORTS_FILES}" ]; then
  printf >&2 'goimports failed for the following files:\n%s\n\nplease run "gofmt -w ." on your changes before committing.\n' "${GOIMPORTS_FILES}"
  exit 1
fi
