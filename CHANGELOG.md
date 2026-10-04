# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- `BSG126` also points a script that loads dybatpho at `dybatpho::split`, which
  takes a delimiter of several characters as it is and keeps empty fields.

- `BSG056` also points a script that loads dybatpho at `dybatpho::curl_do`, which turns an HTTP error into a non-zero status.

- `BSG107` also points a script that loads dybatpho at `dybatpho::curl_timeout`, which sets both limits for one request.

- `BSG117` also points a script that loads dybatpho at `dybatpho::retry` and `dybatpho::curl_do`, which back off between attempts and stop.

- `BSG118` also points a script that loads dybatpho at `dybatpho::run_with_timeout`, which bounds the call and returns 124 on a timeout.

- `BSG081` also points a script that loads dybatpho at `dybatpho::curl_auth_bearer` and `dybatpho::secret_register`, which keep a token off the command line and out of the logs.

## [0.6.0]

### Added

- `BSG110` reports a function, or an entrypoint, whose last statement is
  `test && action`: when the test is false the list fails, that status becomes
  the function's, and `set -e` stops a caller whose call did everything it was
  asked. A predicate built of tests only, such as `[[ -n $1 ]] && dybatpho::is
  file "$1"`, is left alone, as its status is the answer.

- `BSG111` warns about an `EXIT` handler, written inline or as a function of
  the run, that ends with `exit` and a literal status, which replaces the
  status of the script.

- `BSG112` warns about `wait "${pid}"` run as a plain statement in a loop under
  `set -e`: the first job that failed stops the loop, and the others are
  neither waited for nor reported.

- `BSG113` reports `name=value` or `name+=value` on a name the same function,
  or the top level of the file, declares as an array or assigns a list to. It
  writes element 0 only. Declarations and namerefs are left alone.

- `BSG114` reports a `PATH` assignment that holds `.`, an empty element or a
  world-writable directory such as `/tmp`. `${PATH:+:${PATH}}` is understood.

- `BSG115` warns about `.` or `source` of a configuration file under the home
  or configuration directory, the current directory, or a world-writable one:
  sourcing runs every line of it with the rights of the script.

- `BSG116` reports a Bash version check written as
  `BASH_VERSINFO[0] >= X && BASH_VERSINFO[1] >= Y`, which refuses a newer major
  with a lower minor, and a text comparison of `BASH_VERSION`.

- `BSG117` warns about a `while` or `until` loop that waits on `curl`, `wget`,
  `ssh` or another network command and never sleeps, and `BSG118` about `ssh`
  or `scp` with no `ConnectTimeout`, outside `timeout`.

- `BSG119` warns about `find -exec ... {} \;` where `{}` is the last argument,
  which `+` runs once for many files.

- `BSG120` warns about a block of several commands, or a loop, a conditional or
  a `case`, whose standard error goes to `/dev/null` as a whole. The exclusive
  creation `( set -C; : > "${name}" ) 2> /dev/null` is left alone.

- `BSG121` warns about a condition that runs a variable as a command, as in
  `if ${force}; then`. A variable named for the command it holds, such as
  `check_cmd` or `handler`, is left alone.

- `BSG122` warns about `> file 2>&1`, which `&>` writes in one operator, and
  `2>&1 > file`, which still sends errors to the terminal.

- `BSG123` warns about a `# shellcheck disable=` directive with no reason,
  neither after it on the same line nor in a comment right above it.

- `BSG124` warns about a call to `seq`: `{1..5}` makes a fixed range, and
  `for ((i = start; i <= end; i++))` one whose bounds are variables, without a
  process.

- `BSG125` reports `ls` whose output the script reads, inside `$(...)` or at the
  head of a pipeline: names with spaces, newlines or glob characters break it.
  `ls` that only shows a listing, and `git ls-files`, are left alone.

- `BSG126` warns about a string split into fields by a process: `echo` or
  `printf` piped into `cut -d ... -f`, or into an awk program that prints one
  field, and such a `cut` fed by a here-string. `IFS=: read -r` does the same in
  the shell. `cut -c` and `cut -b`, which take characters, are left alone.

## [0.5.0]

### Added

- `BSG099` warns about an entrypoint that uses a feature newer than Bash 4.2 —
  a nameref, `local -`, `mapfile -d`, `wait -n`, `inherit_errexit` or
  `${var@Q}` — and never reads `BASH_VERSINFO`. On the Bash 3.2 that macOS
  ships, such a script fails later with a confusing error instead of a message
  naming the version. A script that sources dybatpho is left alone, as the
  library checks the version itself.

- `BSG100` warns about `echo -e`, `echo -n`, and an `echo` whose first argument
  starts with a variable or a command substitution: a value of `-n` then prints
  nothing, and `-e` turns the backslashes of the data into escapes. An `echo`
  of a fixed message, or one that starts with `$#`, `$?`, `$$` or `$!`, which
  always hold a number, is left alone.

- `BSG101` reports a here document opened with `<<-`, which strips leading tabs
  only, and `BSG102` warns about one whose delimiter is unquoted, that expands
  nothing, and that escapes `$` or a backtick by hand instead of quoting the
  delimiter.

- `BSG103` reports `name["key"]=value` with a literal string key on a name that
  no file of the run declares with `-A`. Bash then reads the key as arithmetic
  and stores every such key at index 0. Namerefs are left alone, and so is a
  key that expands a variable, which may hold a number.

- `BSG104` warns about an entrypoint that stops on error and runs a
  substitution of several commands, or of a function of the run, without
  `shopt -s inherit_errexit`: inside `$(...)` errexit is off, so the
  substitution carries on past a failure.

- `BSG105` reports `exit` or `return` with a literal status outside 0–255,
  which is taken modulo 256, or with 126 or 127, which the shell uses for "not
  executable" and "not found".

- `BSG106` warns about a cleanup handler — one that runs `rm`, or a function
  whose name says it cleans — installed on `INT` or `TERM` without an `exit`.
  It replaces the default action, so the script carries on after Ctrl-C and
  exits 0. A handler named by a function of the run is judged by its body.

- `BSG107` warns about a `curl` call with no `--max-time`, outside `timeout`.
  It waits as long as the server keeps the connection open. Options kept in an
  array or a `--config` file may carry the limit, so such calls are left alone.

- `BSG108` warns about `$(cat file)`, which `$(< file)` does without a process,
  and `BSG109` about a `# shellcheck disable=SC1091` directive, which hides the
  sourced file from ShellCheck where `# shellcheck source=` would name it.

### Changed

- The `.shellcheckrc` of this repository keeps SC2155 on, as the guide now
  does: `readonly X="$(cmd)"` and `export X="$(cmd)"` lose the exit status of
  `cmd` just as `local x="$(cmd)"` does. A project that copied the file should
  copy it again.

- The README no longer asks for the reason of a `# shellcheck disable=`
  directive on a line of its own. ShellCheck reads a second `#` on the same
  line as a comment, and the guide asks for the reason there.

- `BSG010` also reports `readonly X="$(cmd)"` and `export X="$(cmd)"`, at the
  top of a file as well as in a function: the guide no longer recommends the
  one-line form for constants, since the status of `readonly` hides a failing
  `cmd` and leaves an empty constant behind. When ShellCheck reports SC2155 on
  the same line, only `BSG010` is printed.

- `BSG088` suggests `((x += 1))` or `x=$((x + 1))`, the increments the guide now
  recommends, and `BSG097` no longer offers `sort -V` for a version that may
  carry a pre-release suffix, which it orders after the release.

### Fixed

- `BSG033` reports its heading as `Formatting > Function Declaration`, where
  the guide keeps it, instead of a chapter the section never belonged to.

- A message with a literal percent sign is printed as written: the advice of
  `BSG056` read `-w '%%{http_code}'` with a doubled sign.

## [0.4.0]

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

- `BSG056` warns about `curl` without `-f`, `--fail` or `--fail-with-body`:
  it exits 0 on a 404 or a 500 and hands back the error page as if it were the
  answer. A call that asks for `-w '%{http_code}'` to check the status itself
  is left alone.

- `BSG095` warns, in a file that offers a dry run (it mentions `DRY_RUN` or
  `dybatpho::dry_run`), about a function that changes something without
  checking `DRY_RUN` or going through `dybatpho::dry_run`: `rm`, `mv`, `cp` to
  a destination, `install`, `ln -s`, `wget`, `curl -o`, a package manager,
  a mutating `systemctl` verb, or `chezmoi apply`.

- `BSG059` warns about a deprecated command — `apt-key`, `egrep`, `fgrep`,
  `which`, `ifconfig`, `tempfile`, `netstat` — and names its replacement:
  a `signed-by=` keyring, `grep -E`, `grep -F`, `command -v`, `ip addr`,
  `mktemp`, `ss`.

- `BSG096` warns, in an entrypoint, about `.` or `source` of a path built from
  a variable or `$(...)` that the file never tests with `-e`, `-f`, `-r` or
  `-s`, when the status of the `.` is not read: without `set -e` a missing
  library only prints an error, and the script runs on without it.

- `BSG097` warns about `[[ a < b ]]`, `[[ a > b ]]` or an arithmetic `<`,
  `>`, `<=`, `>=` on a variable named like a version, or on a literal such as
  `1.10`: as text `1.10` sorts before `1.9`, and arithmetic stops at the first
  dot. An array element such as `BASH_VERSINFO[0]`, `-lt` on a plain count, and
  `==` are left alone.

- `BSG098` warns about `read` that asks the user — no redirection, no `-u`,
  not reading a loop's or a pipeline's input — with no `-t` timeout, in a
  function or script that never checks for a terminal (`[[ -t 0 ]]`,
  `is_tty`, `is_interactive`): in CI, cron or a pipe it waits forever.

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
