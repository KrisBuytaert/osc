OSC="${OSC:-$(git -C "$(dirname "${BATS_TEST_DIRNAME}")" rev-parse --show-toplevel)/osc}"
OSC_CONFIG="${BATS_TMPDIR}/osc-test-config.yaml"
OS_URL="${OS_URL:-http://localhost:19200}"

write_test_config() {
  cat > "${OSC_CONFIG}" <<EOF
endpoint: ${OS_URL}
username: ""
password: ""
insecure: true
EOF
}

osc() {
  "${OSC}" -c "${OSC_CONFIG}" "$@"
}

os_curl() {
  curl -sf "${OS_URL}$1"
}

wait_for_green() {
  local attempts=30
  until os_curl "/_cluster/health?wait_for_status=green&timeout=5s" >/dev/null 2>&1; do
    ((attempts--)) || { echo "Cluster never went green" >&2; return 1; }
    sleep 2
  done
}

wait_for_yellow() {
  local attempts=30
  until os_curl "/_cluster/health?wait_for_status=yellow&timeout=5s" >/dev/null 2>&1; do
    ((attempts--)) || { echo "Cluster never went yellow" >&2; return 1; }
    sleep 2
  done
}

create_index() {
  curl -sf -X PUT "${OS_URL}/$1" \
    -H 'Content-Type: application/json' \
    -d "{\"settings\":{\"number_of_shards\":$2,\"number_of_replicas\":$3}}" >/dev/null
}

delete_index() {
  curl -sf -X DELETE "${OS_URL}/$1" >/dev/null || true
}

reset_allocation() {
  curl -sf -X PUT "${OS_URL}/_cluster/settings" \
    -H 'Content-Type: application/json' \
    -d '{"transient":{"cluster.routing.allocation.enable":"all","cluster.routing.allocation.exclude._name":null,"cluster.routing.allocation.exclude._ip":null}}' >/dev/null
}
