package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionAmbient    = "Environment > Ambient Environment"
	sectionConfigFile = "Environment > Configuration Files"
)

func init() {
	register(
		Rule{
			Code:     "BSG114",
			Section:  sectionAmbient,
			Severity: lint.SeverityError,
			Doc:      "Keep `.`, empty elements and world-writable directories out of `PATH`",
			Check:    checkUnsafePath,
		},
		Rule{
			Code:     "BSG115",
			Section:  sectionConfigFile,
			Severity: lint.SeverityWarning,
			Doc:      "Parse a configuration file the user or a shared directory controls, instead of sourcing it",
			Check:    checkSourcedConfig,
		},
	)
}

// worldWritable lists the directories any user can write to.
var worldWritable = []string{"/tmp", "/var/tmp", "/dev/shm"}

// checkUnsafePath reports an assignment to `PATH` that holds the current
// directory, as `.` or as an empty element, or a world-writable directory:
// a command planted there runs instead of the system one.
func checkUnsafePath(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		assign, ok := node.(*syntax.Assign)
		if !ok || assign.Name == nil || assign.Name.Value != "PATH" || assign.Value == nil || assign.Index != nil {
			return true
		}
		value := wordSource(assign.Value)
		if assign.Append {
			value = "x" + value
		}
		for _, element := range splitPath(value) {
			element = strings.Trim(element, `"'`)
			switch {
			case element == "":
				r.At(assign.Pos(), "`PATH` has an empty element, which means the current directory: a command planted there runs instead of the system one; drop it, or build the value with `${PATH:+:${PATH}}`")
				return true
			case element == "." || element == "./":
				r.At(assign.Pos(), "`PATH` holds `.`, so a command in whatever directory the script runs from wins over the system one; remove it")
				return true
			}
			for _, dir := range worldWritable {
				if element == dir || strings.HasPrefix(element, dir+"/") {
					r.At(assign.Pos(), "`PATH` holds %s, where any user can plant a command; keep world-writable directories out of it", element)
					return true
				}
			}
		}
		return true
	})
}

// splitPath splits the source of a PATH value on the colons that separate its
// elements, leaving alone the ones inside an expansion such as `${PATH:+:...}`.
func splitPath(src string) []string {
	var out []string
	depth, start := 0, 0
	for i := 0; i < len(src); i++ {
		switch {
		case strings.HasPrefix(src[i:], "${") || strings.HasPrefix(src[i:], "$("):
			depth++
			i++
		case (src[i] == '}' || src[i] == ')') && depth > 0:
			depth--
		case src[i] == ':' && depth == 0:
			out = append(out, src[start:i])
			start = i + 1
		}
	}
	return append(out, src[start:])
}

// configPath matches a sourced path that names a configuration file, and
// userControlled one that a user or anybody can write.
var (
	configPath     = regexp.MustCompile(`(\.conf|\.cfg|\.env|\.ini|rc|/config)["']?$`)
	userControlled = regexp.MustCompile(`^["']?(~|\$\{?HOME\b|\$\{?XDG_CONFIG_HOME\b|\$\{?PWD\b|\./)`)
)

// checkSourcedConfig reports `.` or `source` of a configuration file that the
// user, or anybody, can change: sourcing runs every line of it with the rights
// of the script, which for a script run through sudo are root's.
func checkSourcedConfig(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "." && name != "source" || len(call.Args) < 2 {
			return
		}
		arg := call.Args[1]
		if wordLiteral(arg) == "--" && len(call.Args) > 2 {
			arg = call.Args[2]
		}
		src := wordSource(arg)
		inShared := false
		for _, dir := range worldWritable {
			if strings.Contains(src, dir+"/") {
				inShared = true
			}
		}
		if !inShared && !(userControlled.MatchString(src) && configPath.MatchString(src)) {
			return
		}
		r.At(call.Pos(), "sourcing %s runs every line of it with the rights of the script; parse it as data, with `dybatpho::config_load` or a `key=value` loop that accepts only known keys", src)
	})
}
