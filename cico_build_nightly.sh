#!/bin/bash

set -u
set -e
set -x

SCRIPT_DIR=$(cd "$(dirname "$0")"; pwd)
export SCRIPT_DIR

. "${SCRIPT_DIR}"/cico_functions.sh

load_jenkins_vars
install_deps
build
tag_and_push_nightly
