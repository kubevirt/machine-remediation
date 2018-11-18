#!/bin/bash

set -ex

docker build . -t docker.io/alukiano/openssh-client:28
