#!/usr/bin/env bash
# Copyright  observIQ, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eE

cd "$(dirname "$0")"

clean () {
    rm -rf terraform.tfstate*
    rm -rf providers

    docker ps

    docker-compose down --remove-orphans -t 1
    docker-compose rm --force
}
trap clean ERR

start_containers() {
    docker-compose up -d --remove-orphans --build --force-recreate
}

apply () {
    terraform validate

    terraform apply -auto-approve
}

configure () {
    args="--remote-url https://localhost:3001"
    args="${args} --tls-ca /bindplane-ca.crt"
    args="${args} --tls-cert /bindplane.crt"
    args="${args} --tls-key /bindplane.key"

    eval docker exec integration-bindplane-1 /bindplane get config "$args"
    eval docker exec integration-bindplane-1 /bindplane get destination google-test "$args"
    sleep 5
    eval docker exec integration-bindplane-1 /bindplane get agent -o yaml "$args"
    docker exec integration-agent-1 cat /etc/otel/config.yaml
}

destroy () {
    terraform destroy -auto-approve
}

export TF_CLI_CONFIG_FILE=./dev.tfrc

# fail if BINDPLANE_VERSION is not set
if [[ -z $BINDPLANE_VERSION ]]; then
    echo "BINDPLANE_VERSION is not set"
    exit 1
fi

# fail if BINDPLANE_LICENSE is not set
if [[ -z $BINDPLANE_LICENSE ]]; then
    echo "BINDPLANE_LICENSE is not set"
    exit 1
fi
export BINDPLANE_LICENSE

# trim the v prefix if not latest
if [[ $BINDPLANE_VERSION != "latest" ]]; then
    BINDPLANE_VERSION=$(echo $BINDPLANE_VERSION | sed 's/^v//')
fi
export BINDPLANE_VERSION

echo "using BINDPLANE_VERSION: ${BINDPLANE_VERSION}"

start_containers
sleep 10
apply
configure
destroy
clean
