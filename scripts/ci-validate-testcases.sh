#!/bin/bash

echo "Generate testcases"
make generate-testcases

# check differences
cd sszgen/testcases
if [[ `git status --porcelain` ]]; then
  echo "Testcases have not been generated."
  exit 1
else
  # No changes
  echo "Testcases are correct."
fi
