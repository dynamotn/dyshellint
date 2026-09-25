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
# The whole repository, or one file
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
| `--stdin-filename NAME` | Name to report `-` under |
| `--list-rules` | Print every rule and exit |

Exit status is `0` when nothing failed, `1` when the code has findings, and `2`
when the linter itself could not run. A CI job can tell the two apart.

Every finding carries a code: `BSG###` for a rule of the guide, `SC####` for a
ShellCheck finding, and `FMT001` for a formatting difference. Any of them can be
turned off for a run with `--exclude-rules`, and a ShellCheck finding can be
silenced in place with a `# shellcheck disable=SCXXXX` comment that says why.

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
