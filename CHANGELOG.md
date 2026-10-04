# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `BSG014` reports a plain local in a public function that writes to a
  variable its caller names — through a nameref bound to an argument,
  `printf -v`, `read`, `mapfile` or `readarray`, or by handing the name on to a
  function that does, followed across every file of the run. A caller whose
  variable has the same name as the local gets the local instead and loses the
  value. Only the locals declared before the name is last resolved count, and a
  name a `case` arm pins to literal words is the function's own. Prefixing the
  locals with two underscores keeps them apart.

- `BSG015` warns about a plain local in a function that runs code its caller
  passed — `"$@"`, a positional parameter, or a variable named like `command`
  or `handler` filled from the arguments, run as a command or by `eval`, in a
  subshell too — or that hands such code on to a function that does. The code runs in the function's scope, so it can read and
  change any local declared before it. Files are now parsed before any rule
  runs, so a rule can follow a call into the file that defines the function.

- `BSG016` warns about a variable a function sets without `local` in the ways
  `BSG011` does not see: a `for` loop variable, an arithmetic assignment such
  as `((count++))`, and the variable `read`, `mapfile`, `readarray`,
  `printf -v` or `getopts` fills. It leaks out of the function and overwrites a
  caller's variable of the same name. An UPPERCASE name, `declare -g`, or an
  `@set` tag in the function comment marks a variable set on purpose.

- `BSG047` warns about a call inside `$(...)` to a function that can stop the
  script with `dybatpho::die` or `exit`, directly or through any function it
  calls, followed across every file of the run. Inside a substitution the
  refusal only ends the subshell, so the caller carries on, usually with an
  empty value. A plain assignment whose status is read on the spot — in an
  `if`, after `!`, or followed by `|| return`, `|| exit`, `|| dybatpho::die`,
  `|| status=$?` or a `|| { ... }` block — is left alone; a substitution used as
  an argument, in a test or in `local x=$(...)` loses the status entirely and
  is always reported. The argument checks (`dybatpho::expect_args` and its
  kin) are not counted, as they refuse a call written wrong, not its input.

- `BSG048` warns about `<(...)` fed by a function of the run that can fail —
  one that can stop the script, or that has a `return` other than `return 0` —
  including as the first command of a pipeline inside it. Nothing reads the
  status of a process substitution, so a failure looks like empty output.
  External commands such as `find` are left alone.

- `BSG049` warns, in a file that uses dybatpho, about `"${a[@]}"` or
  `"${a[*]}"` of a local array declared empty that may still be empty where it
  is expanded: Bash 4.3, which dybatpho supports, stops a `set -u` script on
  it. An array counts as filled only after an unconditional assignment of at
  least one element that is not itself a list expansion; filling it inside a
  loop or a branch, with `mapfile`, or from `"$@"` does not. The guarded form
  `${a[@]+"${a[@]}"}` and `${#a[@]}` are left alone.

- `BSG054` warns about `FUNCNAME[N]` with a literal `N` of 2 or more. The
  function that far up the stack depends on every call in between, so a helper
  reached through one more call names the wrong function in its error message.
  Passing the public function's own `${FUNCNAME[0]}` down names it reliably.

- Bats test files (`*.bats`) are linted: a directory walk picks them up, they
  are parsed in the Bats dialect, and shfmt formats them as Bats, so
  `dyshellint test/*.bats` no longer stops on `@test`. Only the formatting
  rules and the rules written for tests run on them; the header, shebang and
  layout rules describe scripts and stay off.

- `BSG061` reports, in a `.bats` file, `assert_output`, `refute_output`,
  `assert_stderr` or `refute_stderr` fed a here document or a here string
  without `-`. bats-assert reads its expectation from standard input only when
  given `-`; without it the assertion ignores the input and passes on any
  non-empty output.

- `BSG080` warns about a variable spliced into literal text that `eval`,
  `dybatpho::dry_run` with a single argument, or `bash -c`/`sh -c` will parse
  as code: a space splits the value, and `$(...)` or `;` in it runs. A value
  quoted with `printf %q` or `${x@Q}`, a word that is a single expansion, and
  `dybatpho::dry_run` given separate words are left alone.

- `BSG081` warns about a credential — a variable named like `TOKEN`, `SECRET`,
  `PASSWORD`, `API_KEY` or `WEBHOOK`, or an `Authorization:` header built from
  a variable — in the arguments of `curl`, `wget` or another HTTP client, where
  `ps` shows it to every user, or printed by `echo`, `printf` or a dybatpho
  logger. A value piped to another command, written to a file, captured by
  `$(...)`, or only measured with `${#x}` is left alone, as are names that only
  point at a secret such as `TOKEN_FILE`.

- `BSG082` reports a path built from `$$`, `$BASHPID`, `$RANDOM` or `$PPID`
  and created with `>`, `>>`, `touch` or `mkdir -p`, wherever it lives: another
  user can plant a link at the name in advance, and those all follow it or
  reuse what is there. `mkdir` without `-p`, a function that turns on `set -C`,
  and paths under `/tmp`, which `BSG045` reports, are left alone.

- `BSG083` warns about an option only the GNU tools accept — `date -d`,
  `sed -i` without a suffix, `readlink -f`, `stat -c`, `find -printf`,
  `grep -P`, `xargs -r`, `mktemp --suffix`, `sort -V`, `base64 -w`,
  `cp --reflink` and their long forms, short options inside a cluster too —
  which fails on macOS, the BSDs and BusyBox. A function that probes the
  flavour first (a `--version` call, `uname`, or a name such as `gnu`, `bsd`
  or `busybox`) and a file that declares itself Linux-only are left alone.

- `BSG055` warns, in a library file, about a raw `trap` that replaces or
  clears the EXIT, INT, TERM, HUP or QUIT handler: the calling script's own
  cleanup is lost. A function that saves the handlers first (`trap -p`, a
  traps-save helper, or `dybatpho::trap`) to put them back is left alone, as
  are `trap -p`, `trap -l` and other signals.

- `BSG073` warns, in a file under `set -u` or using dybatpho, about a bare
  `${NAME}` of an UPPERCASE variable that no file of the run sets: it comes
  from the environment, and `set -u` stops the script when it is missing. A
  default (`${NAME:-}`, `${NAME-x}`), `${NAME:=x}`, the variables Bash sets
  itself, `HOME`, `PATH`, and the variables a dybatpho option spec declares are
  left alone.

- `BSG084` warns about a call inside `$(...)` to a function of the run that
  sets a global in its own shell — an assignment to a name it did not declare
  `local`, `declare -g`, a literal `read`, `mapfile` or `printf -v` target, or
  `dybatpho::secret_register`. The substitution runs it in a subshell, so a
  cache it fills or a secret it registers is gone when the subshell ends.

- `BSG085` warns about `for ((i = 0; i < ${#a[@]}; i++))` over an array the
  function did not build itself — a nameref, or one it never declared local.
  A caller's array can have gaps in its indexes, and counting to its length
  then reads elements that do not exist; `"${!a[@]}"` walks the real ones.

- `BSG086` warns about a `printf` format or an `echo` argument that builds
  JSON by putting a raw value between quotes — `"key":"%s"` or
  `"key":"${value}"` — where a `"`, a `\` or a newline in the value breaks the
  document. A function that escapes its values (an escaper, `jq`, `yq` or a
  JSON helper) and numbers printed with `%d` are left alone.

- `BSG087` warns about `(( ))` or `$(( ))` reading a positional parameter, a
  variable filled from one (or bound by `dybatpho::expect_args`), or a `read`
  target, when the function never checks it: arithmetic reads `08` as a bad
  octal number and runs the `$(cmd)` in `a[$(cmd)]`. A `=~` test, a `case` on
  the variable, a validating call (`dybatpho::is int`, a `validate` or
  `expect_int` helper), and a `10#` prefix count as checks.

- `BSG038` reports a top-level `readonly`, `declare -r` or `typeset -r` in a
  library file that has no source guard before it: sourcing the library a
  second time stops on the constant it already declared. A file that returns
  early when it is already loaded, before the declaration, is left alone.

- `BSG062` warns, in a `.bats` file, about `bash -c` or `sh -c` (also behind
  `run`), whose script runs with an empty `BASH_SOURCE` and breaks code that
  reads it under `set -u`, coverage included; and about a test that runs `git`
  when no file of the run clears `GIT_DIR` and its kin, so that run from a git
  hook every git call lands in the repository being committed to.

- `BSG088` reports `((x++))`, `((x--))` or a comma list ending in one, run as
  a statement in a file under `set -e` or using dybatpho: the expression's
  value is the old one, so the step fails when it was 0 and the script ends.
  A step used as a condition, after `!`, or in an `&&`/`||` list, `((++x))`
  and `x=$((x + 1))` are left alone.

- `BSG039` reports `$0` in a library file: sourced, `$0` names the script that
  sourced it, so `dirname "$0"` finds the wrong directory. The main guard
  `[[ "${BASH_SOURCE[0]}" == "$0" ]]` is left alone.

- `BSG090` warns about `exit` in a function of a library file: it ends the
  whole script that sourced the library, skipping its cleanup and its own error
  handling. A function whose job is to stop (`die`, `fatal`, `abort`), a signal
  or exit handler, a function the file installs with `trap`, and `exit` inside
  a subshell are left alone.

- `BSG091` warns about `cd` or `pushd` in a library function outside a
  `( ... )` subshell when the function never goes back: it moves the whole
  script that sourced the library. A `popd`, `cd -`, `cd "$OLDPWD"`, or a `cd`
  to a variable saved from `$PWD` or `$(pwd)` counts as going back.

- `BSG092` warns about a library function that changes the caller's shell
  state outside a subshell: `set` with `-e`, `-C`, `-f`, `-u`, `-x`, `-a` or
  `-o`, `shopt -s`/`-u`, a bare `IFS=` statement, or `umask` with a mask. A
  function with `local -`, `local IFS`, a saved `$-`, `shopt -p` or
  `$(umask)` to restore from, an `IFS=` in front of one command, and
  `set -- args` are left alone.

- `BSG057` warns about `rm`, `mv`, `cp`, `ln`, `touch`, `mkdir`, `cat`, `ls`,
  `chmod`, `chown`, `grep` or `sed` given an operand that is one scalar
  expansion, such as `"${path}"`, with no `--` before it: a value starting
  with `-` is read as an option. A word with a literal prefix such as
  `"./${path}"`, the mode or owner of `chmod`/`chown`, the value of an option,
  and `grep -e`/`sed -e` patterns are left alone.

- `BSG089` warns about `rm -r`, `chmod -R`, `chown -R`, `chgrp -R` or
  `find -delete` on a path built from a variable the function never checks:
  when it is empty, `rm -rf "${dir}/"` reaches `/`. A `[[ ]]` test of the
  variable, `${dir:?}` in the same word, a variable filled by `mktemp` or a
  temporary-file helper, and one passed to a path-safety check are left alone.

- `BSG093` warns about a `for` loop over a glob whose body never checks that
  the loop variable exists (`-e`, `-f`, `-d`, `-L`, `-r`, `-s`, or
  `dybatpho::is`): without `nullglob`, a glob matching nothing is the loop's
  only word, and the body runs once on a path that does not exist. A file that
  turns on `shopt -s nullglob` or `failglob` is left alone.

- `BSG094` warns about `while read` reading a file, a process substitution or
  a pipe without `|| [[ -n "${line}" ]]` in its condition: `read` fails on a
  last line that has no newline, so the loop drops it. A here document, a here
  string, and `read -d` are left alone.

- `BSG058` reports a `curl` or `wget` download piped into `bash`, `sh` or
  `source /dev/stdin`, or run with `bash <(curl ...)`: whatever the server
  sends runs, and only part of it when the connection drops. Piping into a
  program that reads data, such as `jq`, is left alone.

### Fixed

- A badly formatted file is reported as `FMT001` again when `FORCE_COLOR` is
  set in the environment. shfmt then coloured its diff even into a pipe, the
  hunk headers no longer matched, and every formatting finding was dropped.
  shfmt now runs without the colour-forcing variables and with `NO_COLOR=1`.

- A file shfmt cannot parse no longer stops the whole run with exit code 2.
  The parser of the rules already reports it as `BSG000`, so shfmt skips it and
  goes on with the other files.

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
