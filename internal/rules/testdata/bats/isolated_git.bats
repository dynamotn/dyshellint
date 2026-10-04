setup() {
  load test_helper
  unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE
}

@test "a test that runs git after the helper cleared the hook environment" {
  git init "${BATS_TEST_TMPDIR}/repo"
}
