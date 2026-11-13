package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionMocking = "Testing > Mocking Commands"

func init() {
	register(Rule{
		Code:     "BSG132",
		Section:  sectionMocking,
		Severity: lint.SeverityWarning,
		Doc:      "Fake an external command with a stub on `PATH`, not a shell function of the same name",
		Check:    checkFunctionMock,
		Bats:     true,
	})
}

// mockedCommands are the external commands a test most often fakes. The list
// stays short on purpose: a test helper named `log` or `setup_repo` is no
// stand-in for anything, and only a name that is certainly a real program
// reads as a fake of one.
var mockedCommands = map[string]bool{
	"curl": true, "wget": true, "git": true, "ssh": true, "scp": true, "rsync": true,
	"docker": true, "podman": true, "kubectl": true, "helm": true, "terraform": true,
	"aws": true, "gcloud": true, "az": true, "gh": true, "sudo": true,
	"systemctl": true, "uname": true, "hostname": true, "date": true, "sleep": true,
	"jq": true, "yq": true, "rm": true, "mv": true, "cp": true, "tar": true,
	"make": true, "npm": true, "pip": true, "apt-get": true, "brew": true,
}

// checkFunctionMock reports a test file that fakes an external command by
// defining a function with its name. The function only exists in the test's
// own shell: a child process, `command <name>` and an absolute path all reach
// the real program, and nothing records how it was called. A function that
// calls `command <name>` itself wraps the real program rather than faking it,
// and is left alone.
func checkFunctionMock(f *File, r *Reporter) {
	if !f.Bats {
		return
	}
	eachFunc(f, func(decl *syntax.FuncDecl) {
		name := decl.Name.Value
		if !mockedCommands[name] {
			return
		}
		body := f.Text(decl.Body.Pos(), decl.Body.End())
		if strings.Contains(body, "command "+name) {
			return
		}
		reportWithDybatpho(f, r, decl.Position,
			"`dybatpho::mock_command` puts a recording stub on `PATH`, and `dybatpho::assert_mock_called` checks its calls",
			"the function %q fakes the command only in this shell: a child process, `command %s` and an absolute path still run the real one; put a stub executable first on `PATH` in a temporary directory",
			name, name)
	})
}
