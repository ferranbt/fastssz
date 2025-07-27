#!/bin/bash

echo "Generate testcases"
make build-spec-tests
if [ $? -ne 0 ]; then
  echo "Failed to build spec tests"
  exit 1
fi

make generate-testcases
if [ $? -ne 0 ]; then
  echo "Failed to generate testcases"
  exit 1
fi

# check differences
if [[ `git status --porcelain .` ]]; then
  echo "Codegen has not been generated."
  git diff .
  exit 1
else
  # No changes
  echo "Codegen are correct."
fi
