#!/usr/bin/env bash

SCRIPT_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")"
    pwd
)"

docker build -t alukiano/mro-builder -f ${SCRIPT_DIR}/Dockerfile ${SCRIPT_DIR}
