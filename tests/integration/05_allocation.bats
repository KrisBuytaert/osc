load helpers

# osc-alloc-idx: 1 shard, 3 replicas → 2 replicas UNASSIGNED on 2-node cluster
# Guarantees allocation-explain always has an unassigned shard to explain.
setup_file() {
  write_test_config
  wait_for_yellow
  reset_allocation
  create_index osc-alloc-idx 1 3
  curl -sf "${OS_URL}/_cluster/health?wait_for_status=yellow&timeout=30s" >/dev/null
}

teardown_file() {
  delete_index osc-alloc-idx || true
  reset_allocation
}

@test "allocation-explain: no body exits 0 and explains an unassigned shard" {
  run osc allocation-explain
  [ "$status" -eq 0 ]
  [[ "$output" == *"=== Allocation Explain ==="* ]]
  [[ "$output" == *"shard"* ]]
}

@test "allocation-explain: targeted body explains specific shard" {
  run osc allocation-explain '{"index":"osc-alloc-idx","shard":0,"primary":false}'
  [ "$status" -eq 0 ]
  [[ "$output" == *"osc-alloc-idx"* ]]
  [[ "$output" == *"can_allocate"* ]] || [[ "$output" == *"allocate_explanation"* ]]
}

@test "reroute: empty commands body exits 0 and returns acknowledged" {
  run osc reroute '{"commands":[]}'
  [ "$status" -eq 0 ]
  [[ "$output" == *"acknowledged"* ]]
}

@test "reroute: reads commands body from stdin" {
  run bash -c "echo '{\"commands\":[]}' | \"${OSC}\" -c \"${OSC_CONFIG}\" reroute"
  [ "$status" -eq 0 ]
  [[ "$output" == *"acknowledged"* ]]
}

@test "reroute: exits 1 with no body and no stdin" {
  run bash -c "\"${OSC}\" -c \"${OSC_CONFIG}\" reroute </dev/null"
  [ "$status" -eq 1 ]
}

@test "rebalance: exits 0 after allocation was disabled" {
  curl -sf -X PUT "${OS_URL}/_cluster/settings" \
    -H 'Content-Type: application/json' \
    -d '{"transient":{"cluster.routing.allocation.enable":"none"}}' >/dev/null
  run osc rebalance
  [ "$status" -eq 0 ]
  [[ "$output" == *"cluster.routing.allocation.enable set to all"* ]]
  [[ "$output" == *"Reroute triggered"* ]]
}

@test "rebalance: cluster settings no longer show 'none' after rebalance" {
  curl -sf -X PUT "${OS_URL}/_cluster/settings" \
    -H 'Content-Type: application/json' \
    -d '{"transient":{"cluster.routing.allocation.enable":"none"}}' >/dev/null
  osc rebalance >/dev/null 2>&1
  run osc settings
  [ "$status" -eq 0 ]
  [[ "$output" != *'"enable": "none"'* ]]
}

@test "drain: exits 0 and prints exclusion message" {
  run osc drain os02
  [ "$status" -eq 0 ]
  [[ "$output" == *"os02"* ]]
  [[ "$output" == *"excluded from shard allocation"* ]]
}

@test "drain: cluster settings reflect the exclusion" {
  osc drain os02 >/dev/null 2>&1
  run osc settings
  [ "$status" -eq 0 ]
  [[ "$output" == *"os02"* ]]
}

@test "drain: exits 1 without node argument" {
  run osc drain
  [ "$status" -eq 1 ]
}

@test "undrain: exits 0 and prints exclusions cleared message" {
  osc drain os02 >/dev/null 2>&1
  run osc undrain
  [ "$status" -eq 0 ]
  [[ "$output" == *"Node exclusions cleared"* ]]
}

@test "undrain: cluster settings no longer show os02 as exclusion" {
  osc drain os02 >/dev/null 2>&1
  osc undrain >/dev/null 2>&1
  run osc settings
  [ "$status" -eq 0 ]
  [[ "$output" != *'"_name": "os02"'* ]]
}

@test "undrain: idempotent — second call also exits 0" {
  osc undrain >/dev/null 2>&1
  run osc undrain
  [ "$status" -eq 0 ]
}
