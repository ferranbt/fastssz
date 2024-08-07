#!/bin/bash

echo "Generate testcases"
make build-spec-tests

# check differences
cd spectests
if [[ `git status --porcelain .` ]]; then
  echo "Spectests have not been generated."
  git diff .
  exit 1
else
  # No changes
  echo "Spectests are correct."
fi
