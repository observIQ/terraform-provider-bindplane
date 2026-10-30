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
  "bindplane_gateway.gateway:my-gateway"
  "bindplane_processor_bundle.bundle:my-bundle"
  "bindplane_configuration_v2.configuration:my-config-v2"
  "bindplane_configuration_v2.gateway:my-gateway-config"
)

# Resources that must not drift after import, other than rollout,
# which Bindplane does not store.
strict_import_tests=(
  "bindplane_configuration_v2.configuration"
  "bindplane_configuration_v2.gateway"
)

plan_dir="$(mktemp -d)"
trap 'rm -rf "$plan_dir"' EXIT

# plan_json <file> writes the current plan as JSON to file.
plan_json() {
  terraform plan -input=false -out="$plan_dir/tfplan" > /dev/null
  terraform show -json "$plan_dir/tfplan" > "$1"
}

# assert_plan <file> <addr> <changed> fails unless the plan for addr is a
# no-op, or an in-place update whose changed attributes are exactly
# <changed>, given as a JSON array. An empty array requires a no-op.
assert_plan() {
  local file="$1" addr="$2" changed="$3"
  local result
  result=$(jq -c --arg addr "$addr" '
    [.resource_changes[] | select(.address == $addr)]
    | if length != 1 then {error: "resource not found in plan"}
      else .[0].change as $c
        | {
            actions: $c.actions,
            changed: ([($c.before // {} | keys[]), ($c.after // {} | keys[])]
              | unique | map(select($c.before[.] != $c.after[.])))
          }
      end' "$file")

  if ! jq -e --argjson want "$changed" '
      .actions == ["no-op"]
      or ($want != [] and .actions == ["update"] and .changed == $want)' \
      <<< "$result" > /dev/null; then
    echo "unexpected plan for $addr, want no-op or update of $changed: $result"
    exit 1
  fi
  echo "  $addr: $result"
}

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

echo "=== Checking imported resources only change rollout ==="
plan_json "$plan_dir/import.json"
targets=()
for addr in "${strict_import_tests[@]}"; do
  assert_plan "$plan_dir/import.json" "$addr" '["rollout"]'
  targets+=("-target=$addr")
done

echo "=== Applying imported resources and checking for drift ==="
terraform apply -auto-approve -input=false "${targets[@]}"
plan_json "$plan_dir/apply.json"
for addr in "${strict_import_tests[@]}"; do
  assert_plan "$plan_dir/apply.json" "$addr" '[]'
done

echo "Import test passed: all imports succeeded."
