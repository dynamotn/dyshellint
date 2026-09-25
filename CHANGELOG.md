# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0]

### Added

- `dyshellint` checks a shell script against the
  [Bash coding style guide](https://github.com/dynamotn/bash-coding-style) and
  reports everything as one list. A ✔️ SHOULD or ❌ AVOID rule is an error, a
  ⚠️ CONSIDER rule is a warning, and only errors fail a run unless
  `--warnings-as-errors` says otherwise.
- 36 rules the guide has and no general-purpose shell linter can express:
  function namespaces and privacy prefixes, shdoc headers on the file and on
  every function, variable declaration and loop naming, file layout and the
  single entrypoint call, `eval`, pipes into `while`, loops over command output,
  handmade temporary paths, and the dybatpho conventions
  (`expect_args`, the `_spec_*` command line, the common handlers).
- The same run drives ShellCheck with the project's `.shellcheckrc` and shfmt
  with the formatting options of the guide, so one command covers the whole
  check. A missing tool is reported and skipped rather than fatal.
- `--format json` for CI and editors, `--rules` and `--exclude-rules` to pick
  what runs, `--no-warnings`, and `--list-rules` to print every rule with the
  heading of the guide it comes from. Exit status separates "the code has
  findings" (1) from "the linter could not run" (2).
- `-` reads the script from standard input, with `--stdin-filename` carrying the
  real path, so an editor can check a buffer that has not been saved and the
  project's `.shellcheckrc` and `# shellcheck source=` directives still resolve.
- A [nvim-lint](https://github.com/mfussenegger/nvim-lint) linter in
  `editors/nvim`, with the rule code, the tool behind each finding and the
  heading of the guide on every diagnostic.
- `--version`, stamped with the version, the commit, the tree state and the
  build time at release time.
- Release tooling: `make release` resolves the next version from the
  Conventional Commits since the last tag, runs the gates, then tags and pushes;
  the tag starts the workflow that publishes the archives with goreleaser.
