load helpers

# osc-green-idx:  1 shard, 0 replicas → always STARTED
# osc-yellow-idx: 1 shard, 3 replicas → 2 replicas forced UNASSIGNED on a 2-node cluster
setup_file() {
  write_test_config
  wait_for_yellow
  create_index osc-green-idx 1 0
  create_index osc-yellow-idx 1 3
  curl -sf "${OS_URL}/_cluster/health?wait_for_status=yellow&timeout=30s" >/dev/null
}

teardown_file() {
  delete_index osc-green-idx
  delete_index osc-yellow-idx
  reset_allocation
}

@test "shards: exits 0 and prints Shards header" {
  run osc shards
  [ "$status" -eq 0 ]
  [[ "$output" == *"=== Shards ==="* ]]
}

@test "shards: output contains state column" {
  run osc shards
  [ "$status" -eq 0 ]
  [[ "$output" == *"state"* ]]
}

@test "shards: both test indices appear" {
  run osc shards
  [ "$status" -eq 0 ]
  [[ "$output" == *"osc-green-idx"* ]]
  [[ "$output" == *"osc-yellow-idx"* ]]
}

@test "unassigned: exits 0 and prints Unassigned Shards header" {
  run osc unassigned
  [ "$status" -eq 0 ]
  [[ "$output" == *"=== Unassigned Shards ==="* ]]
}

@test "unassigned: shows UNASSIGNED entries for over-replicated index" {
  run osc unassigned
  [ "$status" -eq 0 ]
  [[ "$output" == *"UNASSIGNED"* ]]
  [[ "$output" == *"osc-yellow-idx"* ]]
}

@test "unassigned: does not show STARTED shards" {
  run osc unassigned
  [ "$status" -eq 0 ]
  [[ "$output" != *"STARTED"* ]]
}

@test "unassigned: header line always present" {
  run osc unassigned
  [ "$status" -eq 0 ]
  [[ "$output" == *"index"* ]]
}

@test "recovery: exits 0 and prints Active Recoveries header" {
  run osc recovery
  [ "$status" -eq 0 ]
  [[ "$output" == *"=== Active Recoveries ==="* ]]
}

@test "recovery: does not show done-stage rows" {
  run osc recovery
  [ "$status" -eq 0 ]
  # Check no data line has 'done' in the stage (4th) column
  while IFS= read -r line; do
    read -ra fields <<< "$line"
    if [ "${#fields[@]}" -ge 4 ] && [ "${fields[0]}" != "index" ]; then
      [ "${fields[3]}" != "done" ]
    fi
  done <<< "$output"
}

@test "recovery: header line always present" {
  run osc recovery
  [ "$status" -eq 0 ]
  [[ "$output" == *"index"* ]]
}
