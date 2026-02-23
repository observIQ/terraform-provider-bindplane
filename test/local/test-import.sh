#!/usr/bin/env bash
# Copyright observIQ, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -e

cd "$(dirname "$0")"

if [[ -z "${TF_CLI_CONFIG_FILE}" ]]; then
  echo "TF_CLI_CONFIG_FILE is not set (e.g. ./dev.tfrc)"
  exit 1
fi

echo "=== Terraform init ==="
terraform init -input=false

# Resource address -> import ID (name) for terraform import.
# Must match resources defined in main.tf.
import_tests=(
  "bindplane_source.host:my-host"
  "bindplane_processor.batch:my-batch"
  "bindplane_destination.custom:example-custom"
  "bindplane_connector.routing:log-router"
  "bindplane_extension.pprof:my-pprof"
)

echo "=== Removing import-test resources from state and re-importing by name ==="
for entry in "${import_tests[@]}"; do
  addr="${entry%%:*}"
  name="${entry##*:}"
  echo "  state rm $addr"
  terraform state rm "$addr"
  echo "  import $addr $name"
  terraform import "$addr" "$name"
done

echo "=== Plan (for visibility only; drift does not fail CI) ==="
terraform plan -input=false
echo "Import test passed: all imports succeeded."
