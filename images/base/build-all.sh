#!/bin/bash

set -ex

for dockerfile in "$@"; do
    image_name="$(dirname ${dockerfile} | xargs basename)"
    cp ${dockerfile} Dockerfile.temp
    docker build -f Dockerfile.temp . -t docker.io/alukiano/${image_name}:28
    rm -rf Dockerfile.temp
done
