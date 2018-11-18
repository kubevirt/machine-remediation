#!/bin/bash

set -e

ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i /usr/local/bin/vagrant.key $@
