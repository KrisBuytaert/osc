load helpers

ISM_POLICY_NAME="osc-retention-test-policy"
TEST_INDEX="osc-retention-test-index"

setup_file() {
  write_test_config
  wait_for_yellow
  create_index "${TEST_INDEX}" 1 0
  ISM_AVAILABLE=1
  if ! curl -sf "${OS_URL}/_plugins/_ism/policies" >/dev/null 2>&1; then
    ISM_AVAILABLE=0
    return
  fi
  curl -sf -X PUT "${OS_URL}/_plugins/_ism/policies/${ISM_POLICY_NAME}" \
    -H 'Content-Type: application/json' \
    -d @"$(dirname "$BATS_TEST_FILENAME")/fixtures/ism_policy.json" >/dev/null
}

teardown_file() {
  delete_index "${TEST_INDEX}"
  curl -sf -X DELETE "${OS_URL}/_plugins/_ism/policies/${ISM_POLICY_NAME}" >/dev/null || true
}

@test "retention: exits 0" {
  run osc retention
  [ "$status" -eq 0 ]
}

@test "retention: prints Fault Tolerance section" {
  run osc retention
  [ "$status" -eq 0 ]
  [[ "$output" == *"=== Fault Tolerance ==="* ]]
  [[ "$output" == *"Data nodes in cluster:"* ]]
}

@test "retention: prints Data Retention per Index section" {
  run osc retention
  [ "$status" -eq 0 ]
  [[ "$output" == *"=== Data Retention per Index ==="* ]]
}

@test "retention: lists our test index" {
  run osc retention
  [ "$status" -eq 0 ]
  [[ "$output" == *"${TEST_INDEX}"* ]]
}

@test "retention: flags a 0-replica index as zero node-loss tolerance" {
  run osc retention
  [ "$status" -eq 0 ]
  [[ "$output" == *"tolerates losing 0 node(s)"* ]]
}
