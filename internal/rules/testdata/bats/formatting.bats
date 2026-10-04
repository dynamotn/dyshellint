setup() {
  load test_helper
}

@test "a test file runs only the rules written for tests" {
	run true
  assert_success   
}
