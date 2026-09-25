// Command dyshellint checks shell scripts against the Bash coding style guide
// at https://github.com/dynamotn/bash-coding-style. It orchestrates the whole
// check: the rules that only this guide has, plus shellcheck and shfmt,
// reported as one list.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
	"gitlab.com/dynamo-tools/dyshellint/internal/rules"
	"gitlab.com/dynamo-tools/dyshellint/internal/version"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, errFindings) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "dyshellint:", err)
		os.Exit(2)
	}
}

// errFindings separates "the code has violations" from "the linter broke",
// which are different exit codes for a CI job.
var errFindings = errors.New("style guide violations found")

type options struct {
	format           string
	only             string
	exclude          string
	shellcheckBinary string
	shfmtBinary      string
	noShellcheck     bool
	noShfmt          bool
	noWarnings       bool
	warningsAsErrors bool
	listRules        bool
	showVersion      bool
	stdinFilename    string
}

func run(args []string, stdout, stderr *os.File) error {
	var opts options
	flags := flag.NewFlagSet("dyshellint", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&opts.format, "format", "text", "output format: text or json")
	flags.StringVar(&opts.only, "rules", "", "comma separated list of rule codes to run, empty means all")
	flags.StringVar(&opts.exclude, "exclude-rules", "", "comma separated list of rule codes to skip")
	flags.StringVar(&opts.shellcheckBinary, "shellcheck", "shellcheck", "shellcheck binary to run")
	flags.StringVar(&opts.shfmtBinary, "shfmt", "shfmt", "shfmt binary to run")
	flags.BoolVar(&opts.noShellcheck, "no-shellcheck", false, "skip the shellcheck pass")
	flags.BoolVar(&opts.noShfmt, "no-shfmt", false, "skip the shfmt pass")
	flags.BoolVar(&opts.noWarnings, "no-warnings", false, "report only errors")
	flags.BoolVar(&opts.warningsAsErrors, "warnings-as-errors", false, "fail the run on warnings too")
	flags.BoolVar(&opts.listRules, "list-rules", false, "print every rule and exit")
	flags.BoolVar(&opts.showVersion, "version", false, "print the version and exit")
	flags.StringVar(&opts.stdinFilename, "stdin-filename", "stdin.sh", "name to report findings under when reading `-`")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: dyshellint [options] <path>...")
		fmt.Fprintln(stderr, "       dyshellint [options] --stdin-filename <name> -")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return err
	}
	if opts.showVersion {
		_, err := fmt.Fprintln(stdout, version.GetBuildInfo())
		return err
	}
	if opts.listRules {
		return listRules(stdout)
	}
	paths := flags.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}

	sources, err := collect(paths, opts, os.Stdin)
	if err != nil {
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			return fmt.Errorf("%s: %w", pathErr.Path, pathErr.Err)
		}
		return err
	}
	defer func() {
		for _, source := range sources {
			source.Close()
		}
	}()
	if len(sources) == 0 {
		fmt.Fprintln(stderr, "dyshellint: no shell files found")
		return nil
	}

	findings, err := checkAll(sources, opts, stderr)
	if err != nil {
		return err
	}
	if opts.noWarnings {
		findings = onlyErrors(findings)
	}
	report := lint.Report{Findings: findings, Format: opts.format, WarningsAsErrors: opts.warningsAsErrors}
	failed, err := report.Write(stdout)
	if err != nil {
		return err
	}
	if failed {
		return errFindings
	}
	return nil
}

// collect turns the command line into the list of sources to check. A single
// `-` reads the program from standard input, which is how an editor lints a
// buffer that has not been saved.
func collect(paths []string, opts options, stdin *os.File) ([]lint.Source, error) {
	if len(paths) == 1 && paths[0] == "-" {
		source, err := lint.StdinSource(stdin, opts.stdinFilename)
		if err != nil {
			return nil, err
		}
		return []lint.Source{source}, nil
	}
	files, err := lint.Discover(paths)
	if err != nil {
		return nil, err
	}
	sources := make([]lint.Source, 0, len(files))
	for _, path := range files {
		source, err := lint.FileSource(path)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}

func checkAll(sources []lint.Source, opts options, stderr *os.File) ([]lint.Finding, error) {
	selected, err := selectRules(opts.only, opts.exclude)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(sources))
	var findings []lint.Finding
	for _, source := range sources {
		paths = append(paths, source.Disk)
		file, err := rules.NewFile(source.Name, source.Content, source.Executable)
		if err != nil {
			// A file the parser rejects is a syntax error, which is a finding of
			// its own rather than a reason to stop the run.
			findings = append(findings, parseFinding(source.Name, err))
			continue
		}
		file.ModeKnown = source.ModeKnown
		findings = append(findings, rules.Run(file, selected)...)
	}
	if !opts.noShellcheck {
		checker := lint.ShellCheck{Binary: opts.shellcheckBinary}
		// A buffer read from standard input lives in a temporary directory, so
		// the configuration of the repository it belongs to has to be named.
		if len(sources) == 1 && sources[0].Disk != sources[0].Name {
			checker.SourceDir = sources[0].ConfigDir()
			checker.RCFile = lint.FindRCFile(checker.SourceDir)
		}
		external, err := checker.Run(paths)
		if err != nil {
			if !errors.Is(err, lint.ErrToolMissing) {
				return nil, err
			}
			fmt.Fprintf(stderr, "dyshellint: skipping shellcheck, %v\n", err)
		}
		findings = append(findings, filterExternal(rename(sources, external), opts)...)
	}
	if !opts.noShfmt {
		external, err := lint.Shfmt{Binary: opts.shfmtBinary}.Run(paths)
		if err != nil {
			if !errors.Is(err, lint.ErrToolMissing) {
				return nil, err
			}
			fmt.Fprintf(stderr, "dyshellint: skipping shfmt, %v\n", err)
		}
		findings = append(findings, filterExternal(rename(sources, external), opts)...)
	}
	return findings, nil
}

// rename puts the findings of the external tools back under the name the
// caller knows, which differs from the path on disk for standard input.
func rename(sources []lint.Source, findings []lint.Finding) []lint.Finding {
	for _, source := range sources {
		findings = source.Rename(findings)
	}
	return findings
}

// parseFinding turns a parse error into a finding, so a broken file is
// reported in the same list as everything else.
func parseFinding(path string, err error) lint.Finding {
	return lint.Finding{
		File:     path,
		Line:     1,
		Column:   1,
		Rule:     "BSG000",
		Severity: lint.SeverityError,
		Level:    lint.SeverityError.String(),
		Message:  fmt.Sprintf("cannot parse as bash: %v", err),
		Source:   "dyshellint",
	}
}

// selectRules applies --rules and --exclude-rules to the registry.
func selectRules(only, exclude string) ([]rules.Rule, error) {
	keep := codeSet(only)
	drop := codeSet(exclude)
	var selected []rules.Rule
	known := map[string]bool{}
	for _, rule := range rules.All() {
		known[rule.Code] = true
		if len(keep) > 0 && !keep[rule.Code] {
			continue
		}
		if drop[rule.Code] {
			continue
		}
		selected = append(selected, rule)
	}
	for code := range keep {
		// An external code in --rules is meaningful for shellcheck, so only an
		// unknown BSG code is a mistake worth reporting.
		if strings.HasPrefix(code, "BSG") && !known[code] {
			return nil, fmt.Errorf("unknown rule %q", code)
		}
	}
	return selected, nil
}

func codeSet(list string) map[string]bool {
	set := map[string]bool{}
	for _, code := range strings.Split(list, ",") {
		if code = strings.TrimSpace(code); code != "" {
			set[strings.ToUpper(code)] = true
		}
	}
	return set
}

// filterExternal applies the same rule selection to shellcheck and shfmt.
func filterExternal(findings []lint.Finding, opts options) []lint.Finding {
	keep := codeSet(opts.only)
	drop := codeSet(opts.exclude)
	var out []lint.Finding
	for _, finding := range findings {
		// A --rules list that names only dyshellint codes is a request for those
		// rules alone, so the external tools fall silent.
		if len(keep) > 0 && !keep[finding.Rule] {
			continue
		}
		if drop[finding.Rule] {
			continue
		}
		out = append(out, finding)
	}
	return out
}

func onlyErrors(findings []lint.Finding) []lint.Finding {
	var out []lint.Finding
	for _, finding := range findings {
		if finding.Severity == lint.SeverityError {
			out = append(out, finding)
		}
	}
	return out
}

func listRules(stdout *os.File) error {
	all := rules.All()
	sort.SliceStable(all, func(i, j int) bool { return all[i].Code < all[j].Code })
	for _, rule := range all {
		if _, err := fmt.Fprintf(stdout, "%s  %-7s  %s\n           %s\n", rule.Code, rule.Severity, rule.Doc, rule.Section); err != nil {
			return err
		}
	}
	return nil
}
