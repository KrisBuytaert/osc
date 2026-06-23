load helpers

setup_file() {
  write_test_config
  wait_for_yellow
}

@test "nodes: exits 0 and prints Nodes header" {
  run osc nodes
  [ "$status" -eq 0 ]
  [[ "$output" == *"=== Nodes ==="* ]]
}

@test "nodes: output contains ip column header" {
  run osc nodes
  [ "$status" -eq 0 ]
  [[ "$output" == *"ip"* ]]
}

@test "nodes: at least 2 node entries present" {
  run osc nodes
  [ "$status" -eq 0 ]
  local node_lines
  node_lines=$(echo "$output" | grep -c "os0" || true)
  [ "$node_lines" -ge 2 ]
}

@test "versions: exits 0 and prints Node Versions header" {
  run osc versions
  [ "$status" -eq 0 ]
  [[ "$output" == *"=== Node Versions ==="* ]]
}

@test "versions: output contains version column" {
  run osc versions
  [ "$status" -eq 0 ]
  [[ "$output" == *"version"* ]]
}

@test "versions: at least 2 nodes shown" {
  run osc versions
  [ "$status" -eq 0 ]
  local node_lines
  node_lines=$(echo "$output" | grep -c "os0" || true)
  [ "$node_lines" -ge 2 ]
}

@test "disk: exits 0 and prints Disk Usage header" {
  run osc disk
  [ "$status" -eq 0 ]
  [[ "$output" == *"=== Disk Usage per Node ==="* ]]
}

@test "disk: output contains disk column headers" {
  run osc disk
  [ "$status" -eq 0 ]
  [[ "$output" == *"disk"* ]]
}

@test "disk: at least 2 nodes shown" {
  run osc disk
  [ "$status" -eq 0 ]
  local node_lines
  node_lines=$(echo "$output" | grep -c "os0" || true)
  [ "$node_lines" -ge 2 ]
}
