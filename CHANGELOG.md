# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.2]

### Added

- `--jobs N` sets how many files ShellCheck and shfmt check at once. It
  defaults to one per CPU.

### Changed

- Linting a directory with many files is several times faster. ShellCheck and
  shfmt now check the files in parallel, largest first, rather than in one
  ShellCheck process on a single core. On the dybatpho library a run drops from
  about 45 seconds to about 7.
- A directory is walked without the files git ignores there, so a generated
  build such as a `dist/` bundle is no longer linted, and no longer floods the
  report. A file named on the command line is still checked even when it is
  ignored.

## [0.3.1]

### Fixed

- An shdoc tag whose text starts on the line below it, as in a bare
  `# @description` followed by an indented paragraph, now counts as documented.
  `BSG020` and `BSG021` no longer report such a file header or function as
  missing its `@description`. A tag with no text at all is still reported.

## [0.3.0]

### Added

- A file header can declare the namespace its functions belong to, with
  `# @namespace NAME` or `# dyshellint namespace=NAME`, so a library no longer
  has to be named after the namespace it exports. `BSG004` measures every
  function against the declaration, and reports a declaration that is not a
  usable name. An entrypoint that declares one is held to it too, instead of to
  the `_` privacy prefix. `BSG060` still looks a test up by the name of the
  file.

## [0.2.0]

### Added

- A `# dyshellint disable=CODE,...` comment silences a finding in place, the
  way ShellCheck's own `# shellcheck disable=` does, and covers every code the
  linter reports: `BSG###`, `SC####` and `FMT001`. At the top of a file it
  covers the whole file, on its own line above a command it covers that command
  and everything nested in it, and at the end of a line it covers that line
  alone. `disable=all` silences every code in scope, and the text after the
  codes is free, for the reason behind the exception.
- A `# shellcheck disable=SC####` comment silences those ShellCheck findings in
  `dyshellint` too, scoped the same way — including at the end of a line, where
  ShellCheck itself ignores the comment. It never silences a `BSG###` or
  `FMT001`, not even through `disable=all`.

## [0.1.1]

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
