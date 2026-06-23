load helpers

setup_file() {
  write_test_config
  wait_for_yellow
}

@test "health: exits 0 and prints cluster_name" {
  run osc health
  [ "$status" -eq 0 ]
  [[ "$output" == *"cluster_name"* ]]
  [[ "$output" == *"osc-test-cluster"* ]]
}

@test "health: output contains number_of_nodes >= 1" {
  run osc health
  [ "$status" -eq 0 ]
  [[ "$output" == *"number_of_nodes"* ]]
}

@test "health: cluster status is green or yellow" {
  run osc health
  [ "$status" -eq 0 ]
  [[ "$output" =~ \"status\" ]]
  [[ "$output" == *"green"* ]] || [[ "$output" == *"yellow"* ]]
}

@test "status: exits 0 and prints Cluster Status header" {
  run osc status
  [ "$status" -eq 0 ]
  [[ "$output" == *"Cluster Status"* ]]
  [[ "$output" == *"indices"* ]]
}

@test "settings: exits 0 and prints Cluster Settings header" {
  run osc settings
  [ "$status" -eq 0 ]
  [[ "$output" == *"Cluster Settings"* ]]
}

@test "settings: output contains persistent and transient keys" {
  run osc settings
  [ "$status" -eq 0 ]
  [[ "$output" == *"persistent"* ]]
  [[ "$output" == *"transient"* ]]
}

@test "get: /_cluster/health returns JSON with cluster_name" {
  run osc get /_cluster/health
  [ "$status" -eq 0 ]
  [[ "$output" == *"cluster_name"* ]]
}

@test "get: exits 1 without path argument" {
  run osc get
  [ "$status" -eq 1 ]
}

@test "get: non-existent path exits non-zero" {
  run osc get /nonexistent_index_xyz
  [ "$status" -ne 0 ]
}

@test "post: raw POST with body returns acknowledged" {
  run osc post /_cluster/reroute '{"commands":[]}'
  [ "$status" -eq 0 ]
  [[ "$output" == *"acknowledged"* ]]
}

@test "post: exits 1 without path argument" {
  run osc post
  [ "$status" -eq 1 ]
}

@test "post: reads body from stdin" {
  run bash -c "echo '{\"commands\":[]}' | \"${OSC}\" -c \"${OSC_CONFIG}\" post /_cluster/reroute"
  [ "$status" -eq 0 ]
  [[ "$output" == *"acknowledged"* ]]
}

@test "put: creates index via raw put" {
  curl -sf -X DELETE "${OS_URL}/osc-put-test" >/dev/null || true
  run osc put /osc-put-test '{"settings":{"number_of_shards":1,"number_of_replicas":0}}'
  [ "$status" -eq 0 ]
  [[ "$output" == *"acknowledged"* ]]
  curl -sf -X DELETE "${OS_URL}/osc-put-test" >/dev/null || true
}

@test "put: creates index via raw put with body from stdin" {
  curl -sf -X DELETE "${OS_URL}/osc-put-stdin-test" >/dev/null || true
  run bash -c "echo '{\"settings\":{\"number_of_shards\":1,\"number_of_replicas\":0}}' | \"${OSC}\" -c \"${OSC_CONFIG}\" put /osc-put-stdin-test"
  [ "$status" -eq 0 ]
  [[ "$output" == *"acknowledged"* ]]
  curl -sf -X DELETE "${OS_URL}/osc-put-stdin-test" >/dev/null || true
}

@test "put: exits 1 without path argument" {
  run osc put
  [ "$status" -eq 1 ]
}

@test "--curl flag prints curl command to stderr" {
  run bash -c "\"${OSC}\" -c \"${OSC_CONFIG}\" --curl health 2>&1"
  [ "$status" -eq 0 ]
  [[ "$output" == *"curl -X GET"* ]]
  [[ "$output" == *"_cluster/health"* ]]
}
