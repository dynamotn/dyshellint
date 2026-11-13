package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionContainerEntrypoints = "Environment > Container Entrypoints"

func init() {
	register(Rule{
		Code:     "BSG135",
		Section:  sectionContainerEntrypoints,
		Severity: lint.SeverityWarning,
		Doc:      "End a container entrypoint with `exec`, so the program it starts replaces Bash as PID 1 and receives `SIGTERM`",
		Check:    checkEntrypointExec,
	})
}

// containerEntrypointName matches the names an image gives the script it
// runs first: `entrypoint.sh`, `docker-entrypoint.sh`, `app_entrypoint.sh`.
var containerEntrypointName = regexp.MustCompile(`(^|[-_.])entrypoint$`)

// checkEntrypointExec reports a container entrypoint, a file named like
// `entrypoint.sh` or `docker-entrypoint.sh`, that ends by running a command instead of replacing itself
// with it. Bash stays PID 1, does not pass `SIGTERM` on, and the container is
// killed when the stop timeout runs out. A last call to a function of the
// project is followed into that function; a call to a function the run cannot
// see, or a last statement that is not a simple command, is left alone.
func checkEntrypointExec(f *File, r *Reporter) {
	if !containerEntrypointName.MatchString(strings.ToLower(BaseName(f.Path))) || len(f.Syntax.Stmts) == 0 {
		return
	}
	p := f.project()
	last := f.Syntax.Stmts[len(f.Syntax.Stmts)-1]
	seen := map[*syntax.FuncDecl]bool{}
	for stmt := last; stmt != nil; {
		call, ok := stmt.Cmd.(*syntax.CallExpr)
		if !ok || len(call.Args) == 0 {
			return
		}
		name := callName(call)
		if name == "" {
			// `"$@"` runs the command the container was given, and is the
			// classic case of a missing `exec`.
			if src := wordSource(call.Args[0]); src == `"$@"` || src == `"${@}"` {
				r.At(call.Pos(), "this entrypoint ends by running the container's command under Bash, which stays PID 1 and never passes `SIGTERM` on; write `exec \"$@\"`")
			}
			return
		}
		switch name {
		case "exec", "exit", "return":
			return
		}
		decl := p.Func(f, name)
		if decl == nil {
			if strings.Contains(name, "::") {
				// A library function the run did not load may well end in `exec`.
				return
			}
			r.At(call.Pos(), "this entrypoint ends by running %q, so Bash stays in front of it as PID 1 and never passes `SIGTERM` on; write `exec %s ...` so the program replaces the shell",
				name, name)
			return
		}
		if seen[decl] {
			return
		}
		seen[decl] = true
		stmt = lastStmt(decl.Body)
	}
}

// lastStmt returns the last statement of a function body written as a block,
// or nil.
func lastStmt(body *syntax.Stmt) *syntax.Stmt {
	block, ok := body.Cmd.(*syntax.Block)
	if !ok || len(block.Stmts) == 0 {
		return nil
	}
	return block.Stmts[len(block.Stmts)-1]
}
