load helpers

setup_file() {
  write_test_config
  wait_for_yellow
  create_index osc-test-idx 1 0
}

teardown_file() {
  delete_index osc-test-idx
}

@test "indices: exits 0 and prints Indices header" {
  run osc indices
  [ "$status" -eq 0 ]
  [[ "$output" == *"=== Indices ==="* ]]
}

@test "indices: output is a JSON array" {
  run osc indices
  [ "$status" -eq 0 ]
  [[ "$output" == *"["* ]]
}

@test "indices: test index appears in list" {
  run osc indices
  [ "$status" -eq 0 ]
  [[ "$output" == *"osc-test-idx"* ]]
}

@test "index: single index exits 0 and shows that index" {
  run osc index osc-test-idx
  [ "$status" -eq 0 ]
  [[ "$output" == *"osc-test-idx"* ]]
}

@test "index: shows pri and rep fields" {
  run osc index osc-test-idx
  [ "$status" -eq 0 ]
  [[ "$output" == *'"pri"'* ]]
  [[ "$output" == *'"rep"'* ]]
}

@test "index: non-existent index exits non-zero" {
  run osc index definitely-does-not-exist-xyz
  [ "$status" -ne 0 ]
}

@test "index: exits 1 without index name argument" {
  run osc index
  [ "$status" -eq 1 ]
}
