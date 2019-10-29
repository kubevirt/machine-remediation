#!/usr/bin/env bash

SCRIPT_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")"
    pwd
)"

docker build -t alukiano/mr-builder -f ${SCRIPT_DIR}/Dockerfile ${SCRIPT_DIR}
