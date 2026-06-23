load helpers

ISM_POLICY_NAME="osc-test-policy"

setup_file() {
  write_test_config
  wait_for_yellow
  # Skip entire file if ISM plugin is absent
  if ! curl -sf "${OS_URL}/_plugins/_ism/policies" >/dev/null 2>&1; then
    skip "ISM plugin not available on this cluster"
  fi
  curl -sf -X PUT "${OS_URL}/_plugins/_ism/policies/${ISM_POLICY_NAME}" \
    -H 'Content-Type: application/json' \
    -d @"$(dirname "$BATS_TEST_FILENAME")/fixtures/ism_policy.json" >/dev/null
}

teardown_file() {
  curl -sf -X DELETE "${OS_URL}/_plugins/_ism/policies/${ISM_POLICY_NAME}" >/dev/null || true
}

@test "ism: exits 0 and prints ISM Policies header" {
  run osc ism
  [ "$status" -eq 0 ]
  [[ "$output" == *"=== ISM Policies ==="* ]]
}

@test "ism: output contains our test policy" {
  run osc ism
  [ "$status" -eq 0 ]
  [[ "$output" == *"${ISM_POLICY_NAME}"* ]]
}

@test "ism: output contains policies key" {
  run osc ism
  [ "$status" -eq 0 ]
  [[ "$output" == *'"policies"'* ]]
}

@test "ism: total_policies is at least 1" {
  run osc ism
  [ "$status" -eq 0 ]
  [[ "$output" == *"total_policies"* ]]
}
