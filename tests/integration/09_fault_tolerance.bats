load helpers

TEST_INDEX="osc-faulttolerance-test-index"

setup_file() {
  write_test_config
  wait_for_yellow
  create_index "${TEST_INDEX}" 1 0
}

teardown_file() {
  delete_index "${TEST_INDEX}"
}

@test "fault-tolerance: exits 0" {
  run osc fault-tolerance
  [ "$status" -eq 0 ]
}

@test "fault-tolerance: prints Fault Tolerance header" {
  run osc fault-tolerance
  [ "$status" -eq 0 ]
  [[ "$output" == *"=== Fault Tolerance ==="* ]]
  [[ "$output" == *"Data nodes in cluster:"* ]]
}

@test "fault-tolerance: does not print Data Retention section" {
  run osc fault-tolerance
  [ "$status" -eq 0 ]
  [[ "$output" != *"=== Data Retention per Index ==="* ]]
}

@test "fault-tolerance: flags a 0-replica index as zero node-loss tolerance" {
  run osc fault-tolerance
  [ "$status" -eq 0 ]
  [[ "$output" == *"tolerates losing 0 node(s)"* ]]
}

@test "fault-tolerance: lists our 0-replica index among least redundant shards" {
  run osc fault-tolerance
  [ "$status" -eq 0 ]
  [[ "$output" == *"${TEST_INDEX}"* ]]
}
