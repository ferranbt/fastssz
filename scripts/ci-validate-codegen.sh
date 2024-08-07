#!/bin/bash

echo "Generate testcases"
make build-spec-tests
make generate-testcases

# check differences
if [[ `git status --porcelain .` ]]; then
  echo "Codegen has not been generated."
  git diff .
  exit 1
else
  # No changes
  echo "Codegen are correct."
fi
