setup() {
  load test_helper
}

@test "a child script started from a string loses its BASH_SOURCE" {
  run bash -c 'printf "%s" "${BASH_SOURCE[0]-}"'
  sh -c 'true'
  run bash -c 'printf "%s" "a@b.com" | dybatpho::ai_redact'
  git init "${BATS_TEST_TMPDIR}/repo"
  git -C "${BATS_TEST_TMPDIR}/repo" status
}

@test "a child script started from a file keeps its BASH_SOURCE" {
  printf 'true\n' > "${BATS_TEST_TMPDIR}/child.sh"
  run bash "${BATS_TEST_TMPDIR}/child.sh"
}
