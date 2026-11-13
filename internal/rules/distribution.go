package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionDistribution = "Environment > Versions and Distribution"

func init() {
	register(
		Rule{
			Code:     "BSG159",
			Section:  sectionDistribution,
			Severity: lint.SeverityWarning,
			Doc:      "Offer `--version` in the spec of an entrypoint",
			Check:    checkSpecVersion,
		},
		Rule{
			Code:     "BSG160",
			Section:  sectionDistribution,
			Severity: lint.SeverityWarning,
			Doc:      "Verify a download against a checksum or a signature before running, installing or unpacking it",
			Check:    checkUnverifiedDownload,
		},
	)
}

// versionCommand matches a `version` subcommand declared in a spec, the other
// way a command line answers which build it is.
var versionCommand = regexp.MustCompile(`dybatpho::opts::cmd\s+["']?version\b`)

// checkSpecVersion reports the root spec of an entrypoint, the one it hands
// to `dybatpho::generate_from_spec` outside any function but `main`, when it
// declares no `--version`. A spec that parses the arguments of a function
// answers to its caller, not to a user asking which build is installed.
func checkSpecVersion(f *File, r *Reporter) {
	if f.Role != RoleEntrypoint {
		return
	}
	up := f.parents()
	roots := map[string]bool{}
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "dybatpho::generate_from_spec" || len(call.Args) < 2 {
			return
		}
		if decl := enclosingFunc(call, up); decl != nil && ownName(f, decl.Name.Value) != "main" &&
			ownName(f, decl.Name.Value) != "_main" {
			return
		}
		roots[wordLiteral(call.Args[1])] = true
	})
	eachFunc(f, func(decl *syntax.FuncDecl) {
		if !roots[decl.Name.Value] {
			return
		}
		body := f.Text(decl.Body.Pos(), decl.Body.End())
		if !strings.Contains(body, "dybatpho::opts::setup") || strings.Contains(body, "--version") ||
			versionCommand.MatchString(body) {
			return
		}
		r.At(decl.Position, "%q declares the command line of the script without `--version`, so a bug report cannot say which build it came from; add `dybatpho::opts::disp \"Show version\" --version action:\"<print the version>\"`, read from a VERSION file or a constant",
			decl.Name.Value)
	})
}

// verification matches a command that checks a file against a digest or a
// signature, or a dybatpho checksum spec such as `sha256:<digest>`.
var verification = regexp.MustCompile(`\b(sha(1|224|256|384|512)sum|shasum|b2sum|gpgv?\b.*--verify|gpgv|cosign\s+verify|minisign|signify|ssh-keygen\s+-Y\s*verify|openssl\s+dgst|verify_checksum|backup_verify)\b|\b(sha1|sha256|sha512):`)

// download is a call that writes what it fetches to a file. target holds the
// variables naming that file, or its literal path; both are empty when the
// server picks the name, as with `curl -O`.
type download struct {
	call   *syntax.CallExpr
	vars   map[string]bool
	path   string
	anyArg bool
}

// checkUnverifiedDownload reports a download that the same file then runs,
// installs or unpacks, when nothing in the file checks a digest or a
// signature. It is a heuristic: the check may live in a helper sourced from
// elsewhere, which is why it only warns.
func checkUnverifiedDownload(f *File, r *Reporter) {
	if verification.Match(f.Src) {
		return
	}
	var downloads []download
	eachCall(f, func(call *syntax.CallExpr, _ string) {
		if d, ok := downloadTarget(call); ok {
			downloads = append(downloads, d)
		}
	})
	reported := map[*syntax.CallExpr]bool{}
	report := func(call *syntax.CallExpr, what string) {
		if reported[call] {
			return
		}
		reported[call] = true
		reportWithDybatpho(f, r, call.Pos(), "`dybatpho::verify_checksum <file> sha256:<digest>` fails on a mismatch", "this download is %s with nothing checking it first, so a truncated file, a compromised mirror or a proxy in between decides what runs; compare it with a published checksum (`sha256sum -c`) or verify its signature (`gpg --verify`, `cosign`, `minisign`) first", what)
	}
	if len(downloads) > 0 {
		eachCall(f, func(call *syntax.CallExpr, _ string) {
			args := unwrapRunAs(unwrapCall(call))
			what := consumes(args)
			if what == "" {
				return
			}
			for _, d := range downloads {
				if d.call.Pos().Offset() < call.Pos().Offset() && d.usedBy(args[1:]) {
					report(d.call, what)
				}
			}
		})
	}
	// `curl ... | tar -x` unpacks a stream nothing can check.
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		pipe, ok := node.(*syntax.BinaryCmd)
		if !ok || (pipe.Op != syntax.Pipe && pipe.Op != syntax.PipeAll) {
			return true
		}
		into := firstCall(pipe.Y)
		if into == nil || consumes(unwrapRunAs(unwrapCall(into))) != "unpacked" {
			return true
		}
		for _, call := range leadingCalls(pipe.X) {
			if downloaders[callName(call)] {
				report(call, "unpacked")
			}
		}
		return true
	})
}

// downloadTarget returns the file a download call writes, when it writes one.
func downloadTarget(call *syntax.CallExpr) (download, bool) {
	args := call.Args
	switch callName(call) {
	case "curl":
		if out := optionWord(args[1:], 'o', "--output"); out != nil {
			return target(call, out), wordLiteral(out) != "-" && wordLiteral(out) != "/dev/null"
		}
		if hasFlag(call, 'O') || hasLong(call, "--remote-name") || hasLong(call, "--remote-name-all") {
			return download{call: call, anyArg: true}, true
		}
	case "wget":
		if hasLong(call, "--spider") {
			return download{}, false
		}
		if out := optionWord(args[1:], 'O', "--output-document"); out != nil {
			return target(call, out), wordLiteral(out) != "-" && wordLiteral(out) != "/dev/null"
		}
		return download{call: call, anyArg: true}, true
	case "dybatpho::curl_download", "dybatpho::curl_resume_download":
		if len(args) > 2 {
			return target(call, args[2]), true
		}
	}
	return download{}, false
}

func target(call *syntax.CallExpr, word *syntax.Word) download {
	d := download{call: call, vars: paramNames(word)}
	if len(d.vars) == 0 {
		d.path = wordLiteral(word)
		d.anyArg = d.path == ""
	}
	return d
}

// optionWord returns the value of an option given as `-o value`, `-o'value'`,
// a cluster ending in the letter such as `-fsSLo value`, `--long value` or
// `--long=value`.
func optionWord(args []*syntax.Word, short byte, long string) *syntax.Word {
	next := func(i int) *syntax.Word {
		if i+1 < len(args) {
			return args[i+1]
		}
		return nil
	}
	joined := func(arg *syntax.Word, value string) *syntax.Word {
		return &syntax.Word{Parts: []syntax.WordPart{&syntax.Lit{ValuePos: arg.Pos(), Value: value}}}
	}
	for i, arg := range args {
		src := wordSource(arg)
		switch {
		case src == long:
			return next(i)
		case strings.HasPrefix(src, long+"="):
			return joined(arg, src[len(long)+1:])
		case len(src) < 2 || src[0] != '-' || src[1] == '-':
			continue
		}
		for j := 1; j < len(src) && !notOptionLetter(rune(src[j])); j++ {
			if src[j] == short {
				if j == len(src)-1 {
					return next(i)
				}
				return joined(arg, src[j+1:])
			}
		}
	}
	return nil
}

// paramNames returns the variables a word expands.
func paramNames(word *syntax.Word) map[string]bool {
	names := map[string]bool{}
	syntax.Walk(word, func(node syntax.Node) bool {
		if pe, ok := node.(*syntax.ParamExp); ok && pe.Param != nil {
			names[pe.Param.Value] = true
		}
		return true
	})
	return names
}

// usedBy reports whether a command names the downloaded file among its
// arguments.
func (d download) usedBy(args []*syntax.Word) bool {
	if d.anyArg {
		return true
	}
	for _, arg := range args {
		if d.path != "" && strings.Contains(wordSource(arg), d.path) {
			return true
		}
		for name := range paramNames(arg) {
			if d.vars[name] {
				return true
			}
		}
	}
	return false
}

// octalMode matches a numeric `chmod` mode.
var octalMode = regexp.MustCompile(`^[0-7]{3,4}$`)

// setsExecute reports whether a `chmod` mode sets an execute bit: `+x`,
// `u=rwx`, or an odd digit among the last three of `755`.
func setsExecute(mode string) bool {
	if octalMode.MatchString(mode) {
		return strings.ContainsAny(mode[len(mode)-3:], "1357")
	}
	return strings.ContainsAny(mode, "+=") && strings.Contains(mode, "x")
}

// consumes returns what a command does to a file it is given: run it, make it
// runnable, install or unpack it, or the empty string.
func consumes(args []*syntax.Word) string {
	if len(args) == 0 {
		return ""
	}
	name := wordLiteral(args[0])
	rest := args[1:]
	switch {
	case name == "chmod":
		for _, arg := range rest {
			if lit := wordLiteral(arg); !strings.HasPrefix(lit, "-") {
				if setsExecute(lit) {
					return "made executable"
				}
				return ""
			}
		}
	case name == "install":
		for _, arg := range rest {
			if lit := wordLiteral(arg); lit == "-d" || lit == "--directory" {
				return ""
			}
		}
		return "installed"
	case shellRunners[name] || name == "source" || name == ".":
		for _, arg := range rest {
			lit := wordLiteral(arg)
			if lit == "-c" || lit == "-s" {
				return ""
			}
			if !strings.HasPrefix(lit, "-") || lit == "" {
				return "run"
			}
		}
	case name == "tar":
		for i, arg := range rest {
			lit := wordLiteral(arg)
			if lit == "--extract" || lit == "--get" {
				return "unpacked"
			}
			// `-xzf`, or the old form `xzf` as the first argument.
			cluster := strings.TrimPrefix(lit, "-")
			if (strings.HasPrefix(lit, "-") && !strings.HasPrefix(lit, "--") || i == 0) &&
				strings.IndexFunc(cluster, notOptionLetter) < 0 && strings.Contains(cluster, "x") {
				return "unpacked"
			}
		}
	case name == "unzip" || name == "dybatpho::archive_extract" || name == "dybatpho::safe_extract":
		return "unpacked"
	}
	return ""
}
