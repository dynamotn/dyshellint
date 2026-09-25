package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionVariableNames = "Naming Conventions > Variable Names"

func init() {
	register(
		Rule{
			Code:     "BSG010",
			Section:  sectionVariableNames,
			Severity: lint.SeverityError,
			Doc:      "Do not declare and assign from a command substitution on the same line",
			Check:    checkDeclareAndAssign,
		},
		Rule{
			Code:     "BSG011",
			Section:  sectionVariableNames,
			Severity: lint.SeverityError,
			Doc:      "Declare every variable used inside a function with `local`",
			Check:    checkImplicitGlobal,
		},
		Rule{
			Code:     "BSG012",
			Section:  sectionVariableNames,
			Severity: lint.SeverityError,
			Doc:      "Name a loop variable after the collection it walks",
			Check:    checkLoopVariableName,
		},
		Rule{
			Code:     "BSG013",
			Section:  sectionVariableNames,
			Severity: lint.SeverityError,
			Doc:      "Declare array locals with `local -a name=()`",
			Check:    checkArrayDeclaration,
		},
	)
}

// declFlags returns the option words of a declaration, such as the `-a` of
// `local -a names=()`. The parser stores them as naked assignments.
func declFlags(decl *syntax.DeclClause) string {
	var flags strings.Builder
	for _, arg := range decl.Args {
		if arg.Name != nil || arg.Value == nil || !arg.Naked {
			continue
		}
		if lit := wordLiteral(arg.Value); strings.HasPrefix(lit, "-") {
			flags.WriteString(lit)
		}
	}
	return flags.String()
}

// wordLiteral renders a word when it is made of plain literals only, and
// returns the empty string when anything has to be expanded to know its value.
func wordLiteral(word *syntax.Word) string {
	if word == nil {
		return ""
	}
	var b strings.Builder
	for _, part := range word.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			b.WriteString(p.Value)
		case *syntax.SglQuoted:
			b.WriteString(p.Value)
		case *syntax.DblQuoted:
			for _, inner := range p.Parts {
				lit, ok := inner.(*syntax.Lit)
				if !ok {
					return ""
				}
				b.WriteString(lit.Value)
			}
		default:
			return ""
		}
	}
	return b.String()
}

// hasCmdSubst reports whether a value is produced by running a command, which
// is what makes an inline declaration swallow an exit status.
func hasCmdSubst(word *syntax.Word) bool {
	if word == nil {
		return false
	}
	found := false
	syntax.Walk(word, func(node syntax.Node) bool {
		switch node.(type) {
		case *syntax.CmdSubst, *syntax.ProcSubst:
			found = true
		}
		return !found
	})
	return found
}

func checkDeclareAndAssign(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		decl, ok := node.(*syntax.DeclClause)
		if !ok || decl.Variant == nil {
			return true
		}
		// `readonly SCRIPT_DIR="$(dirname ...)"` at the top of a file is the
		// shape the guide recommends for constants, so only the declarations
		// that shadow an exit status inside a function are flagged.
		switch decl.Variant.Value {
		case "local", "declare", "typeset":
		default:
			return true
		}
		for _, arg := range decl.Args {
			if arg.Name == nil || !hasCmdSubst(arg.Value) {
				continue
			}
			r.At(arg.Pos(), "`%s %s=$(...)` throws away the exit status of the command; declare `%s` first, then assign it",
				decl.Variant.Value, arg.Name.Value, arg.Name.Value)
		}
		return true
	})
}

func checkImplicitGlobal(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		declared := map[string]bool{}
		syntax.Walk(decl.Body, func(node syntax.Node) bool {
			clause, ok := node.(*syntax.DeclClause)
			if !ok {
				return true
			}
			for _, arg := range clause.Args {
				if arg.Name != nil {
					declared[arg.Name.Value] = true
				}
			}
			return true
		})
		syntax.Walk(decl.Body, func(node syntax.Node) bool {
			// A declaration carries its own assignments; they are handled above.
			if _, ok := node.(*syntax.DeclClause); ok {
				return false
			}
			assign, ok := node.(*syntax.Assign)
			if !ok || assign.Name == nil {
				return true
			}
			name := assign.Name.Value
			// UPPERCASE is the guide's marker for a value the caller sets, so an
			// assignment to one is deliberate rather than an accidental global.
			if declared[name] || name == strings.ToUpper(name) {
				return true
			}
			r.At(assign.Pos(), "%q is assigned inside %q without `local`; it leaks into every function called afterwards",
				name, decl.Name.Value)
			return true
		})
	})
}

// vagueLoopNames are the placeholder names the guide calls out: they say
// nothing about what the loop walks.
var vagueLoopNames = map[string]bool{
	"i": true, "j": true, "k": true, "n": true, "x": true, "y": true, "z": true,
	"item": true, "element": true, "elem": true, "val": true, "value": true,
	"tmp": true, "temp": true, "var": true, "f": true, "s": true, "t": true,
}

func checkLoopVariableName(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		clause, ok := node.(*syntax.ForClause)
		if !ok {
			return true
		}
		iter, ok := clause.Loop.(*syntax.WordIter)
		if !ok || iter.Name == nil {
			return true
		}
		name := strings.TrimLeft(iter.Name.Value, "_")
		if !vagueLoopNames[name] {
			return true
		}
		r.At(iter.Name.Pos(), "loop variable %q says nothing; name it after one element of the collection, as in `for tool in \"${tools[@]}\"`",
			iter.Name.Value)
		return true
	})
}

func checkArrayDeclaration(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		decl, ok := node.(*syntax.DeclClause)
		if !ok || decl.Variant == nil {
			return true
		}
		if decl.Variant.Value != "local" && decl.Variant.Value != "declare" && decl.Variant.Value != "typeset" {
			return true
		}
		flags := declFlags(decl)
		if strings.ContainsAny(flags, "aA") {
			return true
		}
		for _, arg := range decl.Args {
			if arg.Name == nil || arg.Array == nil {
				continue
			}
			r.At(arg.Pos(), "`%s %s=()` declares an array without `-a`; write `%s -a %s=()`",
				decl.Variant.Value, arg.Name.Value, decl.Variant.Value, arg.Name.Value)
		}
		return true
	})
}
