load helpers

setup_file() {
  write_test_config
}

@test "generate-config: creates config file at specified path" {
  local tmpfile="${BATS_TMPDIR}/generated-osc-config-$$.yaml"
  rm -f "${tmpfile}"
  run "${OSC}" -c "${tmpfile}" generate-config
  [ "$status" -eq 0 ]
  [ -f "${tmpfile}" ]
  [[ "$(cat "${tmpfile}")" == *"endpoint"* ]]
  rm -f "${tmpfile}"
}

@test "generate-config: exits 1 if config file already exists" {
  run "${OSC}" -c "${OSC_CONFIG}" generate-config
  [ "$status" -eq 1 ]
}

@test "generate-config: written file has mode 0600" {
  local tmpfile="${BATS_TMPDIR}/generated-osc-perms-$$.yaml"
  rm -f "${tmpfile}"
  "${OSC}" -c "${tmpfile}" generate-config >/dev/null
  local perms
  perms=$(stat -c "%a" "${tmpfile}")
  [ "${perms}" = "600" ]
  rm -f "${tmpfile}"
}

@test "unknown command: exits non-zero" {
  run osc bogus-nonexistent-command
  [ "$status" -ne 0 ]
}

@test "unknown command: prints Unknown command message" {
  run bash -c "\"${OSC}\" -c \"${OSC_CONFIG}\" bogus-command 2>&1"
  [[ "$output" == *"Unknown command"* ]]
}

@test "no command: exits 1 and shows usage" {
  run osc
  [ "$status" -eq 1 ]
  [[ "$output" == *"Usage"* ]] || [[ "$output" == *"Commands"* ]]
}

@test "explicit bad config path: exits 1 with error message" {
  run "${OSC}" -c /nonexistent/path/config.yaml health
  [ "$status" -eq 1 ]
}

@test "-e flag overrides endpoint — unreachable endpoint exits non-zero" {
  run "${OSC}" -c "${OSC_CONFIG}" -e http://127.0.0.1:19999 health
  [ "$status" -ne 0 ]
}
