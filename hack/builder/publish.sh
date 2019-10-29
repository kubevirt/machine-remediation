#!/usr/bin/env bash

SCRIPT_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")"
    pwd
)"

docker tag alukiano/mr-builder docker.io/alukiano/mr-builder
docker push docker.io/alukiano/mr-builder
