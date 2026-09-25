package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionFileHeader   = "Comments > File Header"
	sectionFuncComments = "Comments > Function Comments"
	sectionTodoComments = "Comments > TODO Comments"
	sectionWhichShell   = "Background > Which Shell to Use"
)

func init() {
	register(
		Rule{
			Code:     "BSG020",
			Section:  sectionFileHeader,
			Severity: lint.SeverityError,
			Doc:      "Open every file with an shdoc header carrying @file, @brief and @description",
			Check:    checkFileHeader,
		},
		Rule{
			Code:     "BSG021",
			Section:  sectionFuncComments,
			Severity: lint.SeverityError,
			Doc:      "Document every function with an shdoc comment carrying @description",
			Check:    checkFunctionDescription,
		},
		Rule{
			Code:     "BSG022",
			Section:  sectionFuncComments,
			Severity: lint.SeverityWarning,
			Doc:      "Document the arguments of a function with @arg, or state @noargs",
			Check:    checkFunctionArgs,
		},
		Rule{
			Code:     "BSG023",
			Section:  sectionTodoComments,
			Severity: lint.SeverityError,
			Doc:      "Do not name the author in a TODO comment",
			Check:    checkTodoAuthor,
		},
		Rule{
			Code:     "BSG024",
			Section:  sectionWhichShell,
			Severity: lint.SeverityError,
			Doc:      "Do not put comments before the shebang line",
			Check:    checkCommentBeforeShebang,
		},
	)
}

// headerTags are the shdoc tags the guide asks a file header to carry.
var headerTags = []string{"@file", "@brief", "@description"}

// fileHeader returns the leading comment block of the file: every comment line
// from the first line that is not the shebang until the first line that is not
// a comment.
func fileHeader(f *File) []string {
	var header []string
	for i, line := range f.Lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 && strings.HasPrefix(trimmed, "#!") {
			continue
		}
		if !strings.HasPrefix(trimmed, "#") {
			break
		}
		header = append(header, trimmed)
	}
	return header
}

func checkFileHeader(f *File, r *Reporter) {
	header := strings.Join(fileHeader(f), "\n")
	line := 1
	if f.Shebang != "" {
		line = 2
	}
	var missing []string
	for _, tag := range headerTags {
		if !strings.Contains(header, tag+" ") {
			missing = append(missing, tag)
		}
	}
	if len(missing) == len(headerTags) && header == "" {
		r.AtLine(line, "no file header; open the file with an shdoc comment carrying %s", strings.Join(headerTags, ", "))
		return
	}
	if len(missing) > 0 {
		r.AtLine(line, "file header is missing %s", strings.Join(missing, ", "))
	}
}

// funcComments pairs every function declaration with the comment block that
// precedes it, which is where its shdoc documentation belongs.
func funcComments(f *File) map[*syntax.FuncDecl][]syntax.Comment {
	out := map[*syntax.FuncDecl][]syntax.Comment{}
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok {
			return true
		}
		decl, ok := stmt.Cmd.(*syntax.FuncDecl)
		if !ok || decl.Name == nil {
			return true
		}
		var block []syntax.Comment
		for _, comment := range stmt.Comments {
			// Comments after the declaration belong to the body, not to the header.
			if comment.Pos().Line() < decl.Position.Line() {
				block = append(block, comment)
			}
		}
		out[decl] = block
		return true
	})
	return out
}

func commentText(block []syntax.Comment) string {
	var b strings.Builder
	for _, comment := range block {
		b.WriteString(comment.Text)
		b.WriteString("\n")
	}
	return b.String()
}

func checkFunctionDescription(f *File, r *Reporter) {
	for decl, block := range funcComments(f) {
		if strings.Contains(commentText(block), "@description") {
			continue
		}
		r.At(decl.Position, "%q has no shdoc header; document it with `# @description ...` above the declaration", decl.Name.Value)
	}
}

// positionalRef matches a direct read of a positional parameter, which is what
// an @arg line is expected to describe.
var positionalRef = regexp.MustCompile(`\$\{?[1-9@*]`)

func checkFunctionArgs(f *File, r *Reporter) {
	for decl, block := range funcComments(f) {
		text := commentText(block)
		if text == "" || !strings.Contains(text, "@description") {
			// BSG021 already reports the missing header; one finding is enough.
			continue
		}
		documented := strings.Contains(text, "@arg") || strings.Contains(text, "@noargs")
		if documented {
			continue
		}
		body := f.Text(decl.Body.Pos(), decl.Body.End())
		tag := "@noargs"
		if positionalRef.MatchString(body) || strings.Contains(body, "expect_args") {
			tag = "@arg"
		}
		r.At(decl.Position, "the shdoc header of %q says nothing about its arguments; add `# %s ...`", decl.Name.Value, tag)
	}
}

// todoAuthor matches the `TODO(someone)` form, and a handle left in the text.
var todoAuthor = regexp.MustCompile(`TODO\s*\(([^)]+)\)|TODO\b[^\n]*?(@[A-Za-z][\w.-]+)`)

func checkTodoAuthor(f *File, r *Reporter) {
	eachComment(f, func(comment syntax.Comment) {
		match := todoAuthor.FindStringSubmatch(comment.Text)
		if match == nil {
			return
		}
		name := match[1]
		if name == "" {
			name = match[2]
		}
		r.At(comment.Pos(), "TODO comment names %q; `git blame` already answers who wrote it", strings.TrimSpace(name))
	})
}

// eachComment walks every comment in the file, wherever it is attached.
func eachComment(f *File, fn func(syntax.Comment)) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		if comment, ok := node.(*syntax.Comment); ok {
			fn(*comment)
		}
		return true
	})
	for _, comment := range f.Syntax.Last {
		fn(comment)
	}
}

func checkCommentBeforeShebang(f *File, r *Reporter) {
	if f.Shebang == "" || strings.HasPrefix(strings.TrimSpace(f.Line(1)), "#!") {
		return
	}
	for i, line := range f.Lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#!") {
			r.AtLine(i+1, "the shebang is preceded by %d line(s); `#!` has to be the very first line of the file", i)
			return
		}
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			return
		}
	}
}
