#!/bin/bash -eux

git config --global url."https://${GITHUB_USERNAME}:${GITHUB_TOKEN}@github.com/${GITHUB_REPO_PATH}".insteadOf "https://github.com/${GITHUB_REPO_PATH}"

cwd=$(pwd)

pushd $cwd/dp-cache
  make audit
popd
