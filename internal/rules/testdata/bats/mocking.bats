setup() {
  load test_helper
}

@test "a function named after a command fakes it only in this shell" {
  function curl {
    printf '{"ok":true}\n'
  }
  date() { printf '2026-01-01\n'; }
  run fetch_status
}

@test "a stub on PATH fakes the command for every process" {
  printf '#!/usr/bin/env bash\nprintf "{}\\n"\n' > "${BATS_TEST_TMPDIR}/bin/curl"
  PATH="${BATS_TEST_TMPDIR}/bin:${PATH}"
  run fetch_status
}

@test "a wrapper that calls the real command is no fake" {
  function git {
    command git -c user.name=test "$@"
  }
  function fetch_status {
    printf 'ok\n'
  }
  run fetch_status
}
