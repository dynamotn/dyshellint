package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// Project is what one file cannot know on its own: the functions every file of
// the run defines, and which of them can stop the script or fail. A rule that
// follows a call into another file reads it from here.
type Project struct {
	// funcs maps a function name to its declaration, the last one wins.
	funcs map[string]*syntax.FuncDecl
	// canDie holds the functions that can end the shell they run in, directly
	// with `dybatpho::die` or `exit`, or through a function that can.
	canDie map[string]bool
	// canFail holds the functions that can return a non-zero status: the ones
	// that can die, and the ones with a `return` that is not `return 0`.
	canFail map[string]bool
	// runners holds the functions that run code their caller handed them, and
	// the functions that hand their own caller's code on to one of those.
	runners map[string]bool
}

// dieSinks are the commands that end the shell they run in.
var dieSinks = map[string]bool{"dybatpho::die": true, "exit": true}

// argumentGuards refuse a call the caller wrote wrong, such as a missing
// argument or a variable name that cannot be bound. They are left out of what
// can stop the script: counting them would mark nearly every function, and
// what they refuse is the code of the call, not the input it was handed.
var argumentGuards = map[string]bool{
	"dybatpho::die":                  true,
	"dybatpho::fatal":                true,
	"dybatpho::expect_args":          true,
	"dybatpho::expect_ref":           true,
	"dybatpho::expect_envs":          true,
	"dybatpho::still_has_args":       true,
	"dybatpho::require":              true,
	"__dybatpho_helpers_need_module": true,
}

// Link builds one project out of every file of a run and hands it to each of
// them, so that a call into another file is followed to its definition.
func Link(files []*File) *Project {
	p := &Project{
		funcs:   map[string]*syntax.FuncDecl{},
		canDie:  map[string]bool{},
		canFail: map[string]bool{},
		runners: map[string]bool{},
	}
	for _, f := range files {
		eachFunc(f, func(decl *syntax.FuncDecl) {
			p.funcs[decl.Name.Value] = decl
		})
	}
	p.spread()
	for _, f := range files {
		f.Project = p
	}
	return p
}

// project returns the project the file belongs to, building one out of the
// file alone when no run linked it, as for a buffer an editor pipes in.
func (f *File) project() *Project {
	if f.Project == nil {
		Link([]*File{f})
	}
	return f.Project
}

// CanDie reports whether calling the function can end the shell it runs in.
func (p *Project) CanDie(name string) bool { return p.canDie[name] }

// CanFail reports whether the function can return a non-zero status.
func (p *Project) CanFail(name string) bool { return p.canFail[name] }

// RunsCallerCode reports whether the function runs code its caller passed.
func (p *Project) RunsCallerCode(name string) bool { return p.runners[name] }

// spread works out canDie, canFail and runners to a fixed point over the call
// graph of every function.
func (p *Project) spread() {
	edges := map[string][]string{}
	handsOn := map[string][]string{}
	for name, decl := range p.funcs {
		fed := callerFed(decl)
		ownCalls(decl.Body, func(call *syntax.CallExpr, callee string) {
			switch {
			case dieSinks[callee] && !argumentGuards[name]:
				p.canDie[name] = true
			case callee == "return" && returnsFailure(call):
				p.canFail[name] = true
			case callee == "false":
				p.canFail[name] = true
			}
			if runsCallerCode(call, fed) {
				p.runners[name] = true
			}
			if callee != "" && !argumentGuards[callee] {
				edges[name] = append(edges[name], callee)
				if passesCallerCode(call, fed) {
					handsOn[name] = append(handsOn[name], callee)
				}
			}
		})
	}
	for changed := true; changed; {
		changed = false
		for name, callees := range edges {
			if p.canDie[name] || argumentGuards[name] {
				continue
			}
			for _, callee := range callees {
				if p.canDie[callee] {
					p.canDie[name] = true
					changed = true
					break
				}
			}
		}
		for name, callees := range handsOn {
			if p.runners[name] {
				continue
			}
			for _, callee := range callees {
				if p.runners[callee] {
					p.runners[name] = true
					changed = true
					break
				}
			}
		}
	}
	for name := range p.canDie {
		p.canFail[name] = true
	}
}

// ownCalls walks the simple commands a body runs in its own shell, with the
// literal name of each, or the empty string for a name built by expansion. A command
// inside `$(...)`, `<(...)` or `( ... )` runs in a subshell, where `exit` and
// `dybatpho::die` end only that subshell, so those are left out.
func ownCalls(node syntax.Node, fn func(call *syntax.CallExpr, callee string)) {
	if node == nil {
		return
	}
	syntax.Walk(node, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.CmdSubst, *syntax.ProcSubst, *syntax.Subshell:
			return false
		case *syntax.CallExpr:
			// A command whose name is not a literal, such as `"$@"`, is passed
			// with an empty name: the rules that look for caller code need it.
			if len(n.Args) > 0 {
				fn(n, wordLiteral(n.Args[0]))
			}
		}
		return true
	})
}

// returnsFailure reports whether a `return` can hand back a non-zero status:
// it names a status other than a literal `0`.
func returnsFailure(call *syntax.CallExpr) bool {
	if len(call.Args) < 2 {
		return false
	}
	return wordLiteral(call.Args[1]) != "0"
}

// codeVarPattern matches the names a function gives the code its caller hands
// it: `command`, `handler`, `callback` and the like, plural or prefixed.
var codeVarPattern = map[string]bool{
	"cmd": true, "cmds": true, "command": true, "commands": true,
	"handler": true, "handlers": true, "callback": true, "callbacks": true,
	"producer": true, "spec": true, "action": true, "hook": true,
	"fn": true, "func": true, "function": true, "predicate": true,
}

// isCodeParam reports whether an expansion stands for code the caller passed:
// the positional parameters, or a variable named like `command` or `handler`
// that the function filled from them. A command line the function builds for
// itself, such as `command=(sendmail -t)`, is its own code.
func isCodeParam(pe *syntax.ParamExp, fed map[string]bool) bool {
	if pe == nil || pe.Param == nil || pe.Length || pe.Width || pe.Excl {
		return false
	}
	name := pe.Param.Value
	if isPositionalParam(name) {
		return true
	}
	if !fed[name] {
		return false
	}
	lower := strings.ToLower(strings.TrimLeft(name, "_"))
	for _, part := range strings.Split(lower, "_") {
		if codeVarPattern[part] {
			return true
		}
	}
	return false
}

// isPositionalParam reports whether a parameter name is `@`, `*` or a digit.
func isPositionalParam(name string) bool {
	return name == "@" || name == "*" || len(name) == 1 && name[0] >= '1' && name[0] <= '9'
}

// callerFed returns the variables a function fills from its arguments: those
// assigned a value that expands a positional parameter, and those
// `dybatpho::expect_args` binds.
func callerFed(decl *syntax.FuncDecl) map[string]bool {
	fed := map[string]bool{}
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.FuncDecl:
			return false
		case *syntax.Assign:
			if n.Name != nil && (expandsPositional(n.Value) || n.Array != nil && arrayExpandsPositional(n.Array)) {
				fed[n.Name.Value] = true
			}
		case *syntax.CallExpr:
			if callName(n) == "dybatpho::expect_args" {
				for _, arg := range n.Args[1:] {
					name := wordLiteral(arg)
					if name == "--" {
						break
					}
					fed[name] = true
				}
			}
		}
		return true
	})
	return fed
}

func arrayExpandsPositional(array *syntax.ArrayExpr) bool {
	for _, elem := range array.Elems {
		if expandsPositional(elem.Value) {
			return true
		}
	}
	return false
}

// expandsPositional reports whether a word expands a positional parameter.
func expandsPositional(word *syntax.Word) bool {
	if word == nil {
		return false
	}
	found := false
	syntax.Walk(word, func(node syntax.Node) bool {
		if pe, ok := node.(*syntax.ParamExp); ok && pe.Param != nil && isPositionalParam(pe.Param.Value) {
			found = true
		}
		return !found
	})
	return found
}

// wordParam returns the expansion a word consists of, quoted or not, and nil
// when the word holds anything else.
func wordParam(word *syntax.Word) *syntax.ParamExp {
	if word == nil || len(word.Parts) != 1 {
		return nil
	}
	switch part := word.Parts[0].(type) {
	case *syntax.ParamExp:
		return part
	case *syntax.DblQuoted:
		if len(part.Parts) == 1 {
			if pe, ok := part.Parts[0].(*syntax.ParamExp); ok {
				return pe
			}
		}
	}
	return nil
}

// runsCallerCode reports whether a simple command runs code its caller passed:
// its first word is `"$@"`, a positional parameter or a caller-fed variable
// named like `command`, or it is an `eval` of one of those.
func runsCallerCode(call *syntax.CallExpr, fed map[string]bool) bool {
	if len(call.Args) == 0 {
		return false
	}
	if wordLiteral(call.Args[0]) == "eval" {
		// `eval "$1"` or `eval "${code}"` runs what it was handed; an `eval` of a
		// string the function builds itself is BSG040's business.
		for _, arg := range call.Args[1:] {
			if pe := wordParam(arg); pe != nil && pe.Param != nil && (isPositionalParam(pe.Param.Value) || fed[pe.Param.Value]) {
				return true
			}
		}
		return false
	}
	return isCodeParam(wordParam(call.Args[0]), fed)
}

// passesCallerCode reports whether a call hands the caller's code on: one of its
// arguments is `"$@"` or a caller-fed variable named like `command`.
func passesCallerCode(call *syntax.CallExpr, fed map[string]bool) bool {
	for _, arg := range call.Args[1:] {
		if pe := wordParam(arg); pe != nil && isCodeParam(pe, fed) && !isPositionalDigit(pe) {
			return true
		}
	}
	return false
}

// isPositionalDigit reports whether an expansion is a single positional
// parameter such as `$1`, which is handed on as a value far more often than as
// code to run.
func isPositionalDigit(pe *syntax.ParamExp) bool {
	name := pe.Param.Value
	return len(name) == 1 && name[0] >= '1' && name[0] <= '9'
}

// parents maps every node under root to the node that holds it, which the
// rules need to tell whether a status is read where a command runs.
func parents(root syntax.Node) map[syntax.Node]syntax.Node {
	out := map[syntax.Node]syntax.Node{}
	var stack []syntax.Node
	syntax.Walk(root, func(node syntax.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if len(stack) > 0 {
			out[node] = stack[len(stack)-1]
		}
		stack = append(stack, node)
		return true
	})
	return out
}

// enclosingFunc returns the function a node sits in, or nil at the top level.
func enclosingFunc(node syntax.Node, up map[syntax.Node]syntax.Node) *syntax.FuncDecl {
	for n := up[node]; n != nil; n = up[n] {
		if decl, ok := n.(*syntax.FuncDecl); ok {
			return decl
		}
	}
	return nil
}

// leadingCalls returns the commands a statement starts: the command itself, or
// each command of a pipeline or a `&&`/`||` list.
func leadingCalls(stmt *syntax.Stmt) []*syntax.CallExpr {
	if stmt == nil {
		return nil
	}
	switch cmd := stmt.Cmd.(type) {
	case *syntax.CallExpr:
		return []*syntax.CallExpr{cmd}
	case *syntax.BinaryCmd:
		return append(leadingCalls(cmd.X), leadingCalls(cmd.Y)...)
	}
	return nil
}

// callName returns the literal name of a command, or the empty string.
func callName(call *syntax.CallExpr) string {
	if call == nil || len(call.Args) == 0 {
		return ""
	}
	return wordLiteral(call.Args[0])
}

// funcLocals returns the variables a function declares with `local`, `declare`
// or `typeset`, with the position of their first declaration.
func funcLocals(decl *syntax.FuncDecl) (names []string, at map[string]syntax.Pos) {
	at = map[string]syntax.Pos{}
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		if _, nested := node.(*syntax.FuncDecl); nested {
			// A function declared inside another has locals of its own.
			return false
		}
		clause, ok := node.(*syntax.DeclClause)
		if !ok || clause.Variant == nil {
			return true
		}
		switch clause.Variant.Value {
		case "local", "declare", "typeset":
		default:
			return true
		}
		if strings.Contains(declFlags(clause), "g") {
			return true
		}
		for _, arg := range clause.Args {
			if arg.Name == nil {
				continue
			}
			if _, seen := at[arg.Name.Value]; !seen {
				at[arg.Name.Value] = arg.Pos()
				names = append(names, arg.Name.Value)
			}
		}
		return true
	})
	return names, at
}

// isPrivateName reports whether a variable name carries the private prefix a
// caller's own names never use: two leading underscores.
func isPrivateName(name string) bool {
	return strings.HasPrefix(name, "__")
}

// isPublicFunc reports whether a function is part of a library's interface: its
// name does not start with an underscore.
func isPublicFunc(decl *syntax.FuncDecl) bool {
	return !strings.HasPrefix(decl.Name.Value, "_")
}
