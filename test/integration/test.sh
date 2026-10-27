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
    echo "cleaning up"

    rm -rf terraform.tfstate*
    rm -rf providers

    docker ps

    debug_logs

    docker-compose down --remove-orphans -t 1
    docker-compose rm --force
}
trap clean ERR

start_containers() {
    docker-compose up -d --remove-orphans --build --force-recreate
}

debug_logs() {
    echo "getting container logs"

    docker logs integration-bindplane-1
}

configure() {
    echo "configuring bindplane"

    if [[ $ORGS == "true" ]]; then
        echo "creating organization"

        curl -v \
            --cacert ../../client/tls/bindplane-ca.crt \
            --cert ../../client/tls/bindplane.crt \
            --key ../../client/tls/bindplane.key \
            -u tfu:tfp \
            https://localhost:3100/v1/organizations \
            -X POST \
            -d '{"organizationName": "init", "accountName": "project", "eulaAccepted":true}'| jq .
    else
        echo "creating account"

        curl -v \
            --cacert ../../client/tls/bindplane-ca.crt \
            --cert ../../client/tls/bindplane.crt \
            --key ../../client/tls/bindplane.key \
            -u tfu:tfp \
            https://localhost:3100/v1/accounts \
            -X POST \
            -d '{"displayName": "init"}' | jq .
    fi

    # Allow bindplane to seed the database by making a GET request
    # and waiting 20 seconds.
    curl -v \
        --cacert ../../client/tls/bindplane-ca.crt \
        --cert ../../client/tls/bindplane.crt \
        --key ../../client/tls/bindplane.key \
        -u tfu:tfp \
        https://localhost:3100/v1/source-types \
        -X GET
    sleep 20
}

apply() {
    terraform validate

    terraform apply -auto-approve || debug_logs
}

test_resources() {
    args="--remote-url https://localhost:3001"
    args="${args} --tls-ca /bindplane-ca.crt"
    args="${args} --tls-cert /bindplane.crt"
    args="${args} --tls-key /bindplane.key"

    eval docker exec integration-bindplane-1 /bindplane get config "$args"
    eval docker exec integration-bindplane-1 /bindplane get destination google-test "$args"

    # terraform apply errors are not fatal (see apply), so fail here
    # unless Bindplane returns the values set in main.tf.
    expect_resource destination google-test \
        '.metadata.displayName == "Google Test" and .metadata.description == "my terraform description"'

    # Older versions persist the description, but these configurations
    # are not created on them: they reference processor types that do
    # not exist there (processor_bundle in 1.76.0, parse_timestamp_v2
    # in 1.96.0 and older).
    if version_at_least 1.89.3; then
        expect_resource configuration testtf \
            '.metadata.description == "my terraform configuration description"'
    fi
    if version_at_least 1.104.0; then
        expect_resource configuration my-config-v2 \
            '.metadata.description == "my terraform configuration v2 description"'
    fi

    if [[ $GATEWAY == "true" ]]; then
        test_gateway
    fi
}

# bindplane_api performs a GET request against the Bindplane API.
bindplane_api() {
    curl -sf \
        --cacert ../../client/tls/bindplane-ca.crt \
        --cert ../../client/tls/bindplane.crt \
        --key ../../client/tls/bindplane.key \
        -u tfu:tfp \
        "https://localhost:3100/v1/$1"
}

fail() {
    echo "FAIL: $*"
    return 1
}

# test_gateway fails the test when the gateway or its configuration
# were not applied as expected. Apply errors do not fail the test, so
# this check is what makes the gateway test required.
test_gateway() {
    echo "testing gateway"

    gateway=$(bindplane_api resources/Gateway/my-gateway) || fail "gateway my-gateway not found"
    echo "$gateway" | jq .
    echo "$gateway" | jq -e '.resource.spec.configuration == "my-gateway-config"' > /dev/null \
        || fail "gateway configuration is not my-gateway-config"
    echo "$gateway" | jq -e '.resource.spec.loadBalancerHostname == "gateway.example.com"' > /dev/null \
        || fail "gateway loadBalancerHostname is not gateway.example.com"

    config=$(bindplane_api configurations/my-gateway-config) || fail "configuration my-gateway-config not found"
    echo "$config" | jq -e '.configuration.metadata.labels.role == "gateway"' > /dev/null \
        || fail "configuration my-gateway-config does not have label role=gateway"

    echo "gateway test passed"
}

# version_at_least <version> succeeds when BINDPLANE_VERSION is latest
# or at least <version>.
version_at_least() {
    if [[ $BINDPLANE_VERSION == "latest" ]]; then
        return 0
    fi
    [[ $(printf '%s\n' "$1" "$BINDPLANE_VERSION" | sort -V | head -n1) == "$1" ]]
}

# expect_resource <kind> <name> <jq filter> fails unless the resource
# returned by the Bindplane CLI matches the jq filter.
expect_resource() {
    local kind=$1
    local name=$2
    local filter=$3
    local out

    # The CLI can print its own JSON log lines to stdout, so only
    # objects with metadata (the resource) are checked.
    out=$(eval docker exec integration-bindplane-1 /bindplane get "$kind" "$name" -o json "$args") || true
    if ! echo "$out" | jq -e "select(.metadata != null) | $filter" > /dev/null; then
        echo "FAIL: $kind $name does not match: $filter"
        echo "$out"
        return 1
    fi
    echo "OK: $kind $name matches: $filter"
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
export ORGS=false

if [[ $(printf '%s\n' "$BINDPLANE_VERSION" "1.58.0" | sort -V | head -n1) == "1.58.0" && "$BINDPLANE_VERSION" != "1.58.0" ]]; then
    export ORGS=true
elif [[ "$BINDPLANE_VERSION" == "latest" ]]; then
    export ORGS=true
fi

# Gateways require Bindplane v1.103.0 or newer. They are tested on
# latest and on every version from v1.103.0, and skipped on older versions.
GATEWAY_MIN_VERSION="1.103.0"
export GATEWAY=false
if [[ "$BINDPLANE_VERSION" == "latest" ]]; then
    export GATEWAY=true
    echo "gateway tests enabled: BINDPLANE_VERSION is latest"
elif [[ $(printf '%s\n' "$GATEWAY_MIN_VERSION" "$BINDPLANE_VERSION" | sort -V | head -n1) == "$GATEWAY_MIN_VERSION" ]]; then
    export GATEWAY=true
    echo "gateway tests enabled: ${BINDPLANE_VERSION} is ${GATEWAY_MIN_VERSION} or newer"
else
    echo "gateway tests skipped: ${BINDPLANE_VERSION} is older than ${GATEWAY_MIN_VERSION}"
fi
export TF_VAR_enable_gateway=$GATEWAY

echo "using BINDPLANE_VERSION: ${BINDPLANE_VERSION}"

start_containers
sleep 10
configure
apply
test_resources
destroy
clean
