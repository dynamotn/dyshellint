package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionSecrets = "Features and Bugs > Secrets and Credentials"

func init() {
	register(Rule{
		Code:     "BSG081",
		Section:  sectionSecrets,
		Severity: lint.SeverityWarning,
		Doc:      "Keep secrets out of a command's arguments and out of what a script prints",
		Check:    checkSecretExposure,
	})
}

// secretName matches the names a credential goes by.
var secretName = regexp.MustCompile(`(?i)(token|secret|passw(or)?d|api_?key|webhook|credential)`)

// notSecretName matches the names that only point at a secret, such as the
// file it lives in or the variable it is read from.
var notSecretName = regexp.MustCompile(`(?i)(_|^)(file|path|dir|directory|name|vars?|env|ref|header_name|count|len|length|type|kind|source|mode|id|placeholder|prompt|staging|tokens)$`)

// strongSecretName matches a name that holds a credential whatever context it
// is printed in. A bare `token` is too often a parser's token to count there.
var strongSecretName = regexp.MustCompile(`(?i:secret|passw(or)?d|api_?key|webhook|credential|(api|auth|access|bearer|refresh|session|bot|github|gitlab|forge|private|oauth)_?token)|^[A-Z0-9_]*TOKEN$`)

// httpClients put their arguments on a command line every user of the machine
// can read through `ps`.
var httpClients = map[string]bool{"curl": true, "wget": true, "http": true, "https": true, "xh": true, "httpie": true}

// loggers print what they are given, to the terminal or a log file.
var loggers = map[string]bool{
	"echo": true, "printf": true,
	"dybatpho::print": true, "dybatpho::info": true, "dybatpho::debug": true,
	"dybatpho::warn": true, "dybatpho::error": true, "dybatpho::success": true,
	"dybatpho::fatal": true, "dybatpho::die": true, "dybatpho::header": true,
}

func checkSecretExposure(f *File, r *Reporter) {
	var up map[syntax.Node]syntax.Node
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok {
			return true
		}
		call, ok := stmt.Cmd.(*syntax.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		name := callName(call)
		switch {
		case httpClients[name]:
		case loggers[name]:
			if up == nil {
				up = parents(f.Syntax)
			}
			// A value piped to another command or written to a file is handed
			// over, not shown; that is how a secret should travel.
			if (name == "echo" || name == "printf") && !printsToTerminal(stmt, up) {
				return true
			}
		default:
			return true
		}
		for _, arg := range call.Args[1:] {
			if secret := secretIn(arg, name); secret != "" {
				verb := "puts it on the command line, where `ps` shows it to every user"
				if !httpClients[name] {
					verb = "prints it"
				}
				reportWithDybatpho(f, r, arg.Pos(), "`dybatpho::curl_auth_bearer <url> <token>` sends a token without putting it on a command line, and `dybatpho::secret_register` masks it in what dybatpho logs", "%q looks like a credential and `%s` %s; pass it on stdin or through a file only the user can read",
					secret, name, verb)
				return true
			}
		}
		return true
	})
}

// secretIn returns the credential a word expands, or an `Authorization:`
// header that splices one in.
func secretIn(word *syntax.Word, command string) string {
	var found string
	syntax.Walk(word, func(node syntax.Node) bool {
		if found != "" {
			return false
		}
		switch n := node.(type) {
		case *syntax.CmdSubst:
			return false
		case *syntax.ParamExp:
			if n.Param == nil || n.Length || n.Excl {
				return false
			}
			v := n.Param.Value
			if !secretName.MatchString(v) || notSecretName.MatchString(v) {
				return false
			}
			if httpClients[command] || strongSecretName.MatchString(v) {
				found = v
			}
			return false
		}
		return true
	})
	if found == "" && httpClients[command] {
		text := wordSource(word)
		if strings.Contains(strings.ToLower(text), "authorization:") && strings.Contains(text, "$") {
			found = "Authorization"
		}
	}
	return found
}

// printsToTerminal reports whether a statement writes to the script's own
// output: it is not redirected, not piped into another command, and not run
// inside a command substitution that captures it.
func printsToTerminal(stmt *syntax.Stmt, up map[syntax.Node]syntax.Node) bool {
	for _, redir := range stmt.Redirs {
		switch redir.Op {
		case syntax.RdrOut, syntax.AppOut, syntax.RdrAll, syntax.AppAll, syntax.RdrClob:
			return false
		}
	}
	for n := up[stmt]; n != nil; n = up[n] {
		switch p := n.(type) {
		case *syntax.BinaryCmd:
			if p.Op == syntax.Pipe || p.Op == syntax.PipeAll {
				if syntax.Node(p.X) == syntax.Node(stmt) {
					return false
				}
			}
		case *syntax.CmdSubst, *syntax.ProcSubst:
			return false
		case *syntax.FuncDecl:
			return true
		}
	}
	return true
}
