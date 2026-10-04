# dyshellint

A linter for the [Bash coding style guide](https://github.com/dynamotn/bash-coding-style).

It orchestrates the whole check in one command: the rules that only that guide
has — namespaces, shdoc headers, file layout, the
[dybatpho](https://github.com/dynamotn/dybatpho) conventions — plus
[ShellCheck](https://www.shellcheck.net/) and
[shfmt](https://github.com/mvdan/sh), reported as a single list.

Rules of the guide marked ✔️ SHOULD or ❌ AVOID are reported as errors, the
⚠️ CONSIDER ones as warnings. Only errors fail a run, unless
`--warnings-as-errors` says otherwise.

## Install

```sh
go install gitlab.com/dynamo-tools/dyshellint/cmd/dyshellint@latest
```

ShellCheck and shfmt are optional: when one is missing, `dyshellint` says so on
STDERR and reports everything else.

## Use

```sh
# The whole repository, or one file. A directory is walked without what git
# ignores there, such as a generated bundle; a file named outright is checked
# even when it is ignored.
dyshellint ./scripts
dyshellint ./scripts/deploy.sh

# A buffer that has not been saved, for an editor
cat script.sh | dyshellint --stdin-filename script.sh -

# Machine readable, for CI or an editor
dyshellint --format json ./scripts

# Every rule, with the heading of the guide it comes from
dyshellint --list-rules
```

| Option | Meaning |
| --- | --- |
| `--format text\|json` | Output format, `text` by default |
| `--rules CODE,...` | Run only these rules |
| `--exclude-rules CODE,...` | Skip these rules |
| `--no-warnings` | Report only the ✔️ SHOULD and ❌ AVOID tier |
| `--warnings-as-errors` | Fail the run on ⚠️ CONSIDER findings too |
| `--no-shellcheck`, `--no-shfmt` | Skip an external tool |
| `--shellcheck`, `--shfmt` | Point at another binary |
| `--jobs N` | Files ShellCheck and shfmt check at once, one per CPU by default |
| `--stdin-filename NAME` | Name to report `-` under |
| `--list-rules` | Print every rule and exit |

Exit status is `0` when nothing failed, `1` when the code has findings, and `2`
when the linter itself could not run. A CI job can tell the two apart.

Every finding carries a code: `BSG###` for a rule of the guide, `SC####` for a
ShellCheck finding, and `FMT001` for a formatting difference. Any of them can be
turned off for a run with `--exclude-rules`, or silenced in place with a
disable comment.

Bats test files (`*.bats`) are linted too: they are parsed with the Bats
dialect, formatted by shfmt as Bats, and checked by ShellCheck and by the rules
written for tests, such as strict output assertions. The rules about headers,
shebangs and layout, which describe scripts, do not apply to them.

## Silence a finding in place

A `# dyshellint disable=CODE,...` comment works the way ShellCheck's own
`# shellcheck disable=` does, and covers every code `dyshellint` reports: the
rules of the guide, ShellCheck and shfmt alike. Anything after the codes is free
text, which is where the reason for the exception belongs.

Where the comment sits decides how far it reaches:

```sh
#!/usr/bin/env bash
# dyshellint disable=BSG020 # the header lives in the wrapper script
# At the top of the file, before the first command: the whole file.

set -euo pipefail

# dyshellint disable=BSG002,BSG003 # the name is the public API of a vendored lib
function Demo() {
  # On its own line: the command below and everything nested in it, so a
  # comment above a function covers the whole function.
  echo "hi"
}

grep -r $pattern . # dyshellint disable=SC2086 # the pattern is a word list
# At the end of a line: that line alone.
```

`disable=all` silences every code in the scope of the comment. A file the parser
cannot read has no directives, since its only finding is the syntax error
itself.

A `# shellcheck disable=SC####` comment is read as well, and silences the
ShellCheck codes it names — never a `BSG###` or `FMT001`, not even through
`disable=all`. It is scoped the way every directive here is, so it also works at
the end of a line, where ShellCheck itself ignores it. Give the reason after a
second `#` on the same line, as the guide asks: ShellCheck reads it as a comment.

```sh
# shellcheck disable=SC2034 # read by the sourced template
template_name="release"
```

## Give a file another namespace

`BSG004` expects the functions of a library to be named after their file:
`scripts/lib/package_manager.sh` holds `package_manager::install` and
`__package_manager_helper`. When the file reads better under another name, the
header declares it, in either spelling:

```sh
# @file pacman_wrapper.sh
# @brief Install packages with pacman
# @description ...
# @namespace pkg
# dyshellint namespace=pkg   # the same thing, in the directive vocabulary

function pkg::install {
  :
}
```

The declaration is only read in the file header, before the first command,
which is where the rest of the shdoc tags live. It must be a name a function can
carry — lowercase, digits and underscores — and a declaration that is not is
reported as a `BSG004` of its own, with the file name standing in for it.

A script that carries no `lib/` path can declare one too, which holds its
functions to that namespace instead of asking for the `_` prefix an entrypoint
otherwise takes. `BSG060` is unaffected: a test file is looked up by the name of
the library it covers, so `pacman_wrapper.sh` is still tested by
`test/pacman_wrapper.bats`.

## Configure

`dyshellint` has no configuration file of its own. It reads the `.shellcheckrc`
of the project being checked, and passes the formatting options of the guide to
shfmt explicitly. The [`.shellcheckrc`](.shellcheckrc) and
[`.editorconfig`](.editorconfig) of this repository are the ones the guide asks
for, and are meant to be copied into a project.

## Editors

[`editors/nvim`](editors/nvim) holds a ready-made
[nvim-lint](https://github.com/mfussenegger/nvim-lint) definition. It pipes the
buffer in on standard input, so a script is checked while it is being written
rather than only once it is saved.

## Develop

The rules live in `internal/rules`, one file per chapter of the guide, and each
one is registered with the heading it enforces. A rule is exercised by a fixture
under `internal/rules/testdata`: a `.sh` file and the `.want` file listing what
it should report. After adding a rule or a fixture, regenerate the expectations
and read the diff:

```sh
go test ./internal/rules -update
```

`TestAllRulesAreCovered` fails when a rule has no fixture, so a new rule arrives
with the example that documents it.
