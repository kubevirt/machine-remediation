#!/bin/bash

set -ex

for dockerfile in "$@"; do
    image_name="$(dirname ${dockerfile} | xargs basename)"
    docker push docker.io/alukiano/${image_name}:28
done
