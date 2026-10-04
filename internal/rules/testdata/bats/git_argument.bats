setup() {
  load test_helper
}

@test "a package called git is an argument, not a git call" {
  run install_package git
  run_traced --separate-stderr install_package git
}
