#!/usr/bin/env bash

SCRIPT_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")"
    pwd
)"

docker tag alukiano/mro-builder docker.io/alukiano/mro-builder
docker push docker.io/alukiano/mro-builder
