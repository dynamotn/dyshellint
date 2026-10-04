setup() {
  load test_helper
}

@test "an output assertion fed a here document without - checks nothing" {
  run printf 'one\ntwo\n'
  assert_output << EOF
one
two
EOF
  assert_stderr <<< ""
  refute_output --partial << EOF
three
EOF
}

@test "an output assertion given - compares its here document" {
  run printf 'one\ntwo\n'
  assert_output - << EOF
one
two
EOF
  assert_output --partial - <<< "one"
  assert_output "one"$'\n'"two"
}
