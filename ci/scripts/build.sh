#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-cache
  make build
popd