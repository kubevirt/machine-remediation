#!/usr/bin/env bash

SCRIPT_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")"
    pwd
)"

docker tag alukiano/mrro-builder docker.io/alukiano/mrro-builder
docker push docker.io/alukiano/mrro-builder
