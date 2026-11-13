package rules

import (
	"regexp"
	"sort"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionCommentedCode = "Comments > Commented-out Code"

func init() {
	register(Rule{
		Code:     "BSG144",
		Section:  sectionCommentedCode,
		Severity: lint.SeverityWarning,
		Doc:      "Delete code instead of commenting it out; git keeps the old version",
		Check:    checkCommentedCode,
	})
}

var (
	// codeSignal matches what prose rarely holds and code usually does: a
	// declaration or an assignment, an expansion, `&&`, `||` or a pipe, a
	// test, or a call into a namespace. `;` and `>` are left out: prose uses
	// them too.
	codeSignal = regexp.MustCompile(`^(local|export|readonly|declare|typeset)\s|^[A-Za-z_][A-Za-z0-9_]*(\[[^]]*\])?\+?=|\$[{(A-Za-z_@#?0-9]|\s\|\s|&&|\|\||\[\[\s|^[A-Za-z_][A-Za-z0-9_]*::[A-Za-z_]`)
	// sentenceStart is a capitalised word, which opens a sentence far more
	// often than a line of code.
	sentenceStart = regexp.MustCompile(`^\p{Lu}\p{Ll}`)
	// codeKeyword is a reserved word that opens or closes a compound command.
	codeKeyword = map[string]bool{
		"if": true, "then": true, "elif": true, "else": true, "fi": true, "for": true, "while": true, "until": true,
		"do": true, "done": true, "case": true, "esac": true, "return": true, "function": true, "}": true,
	}
	// plainWord is a word of any language, so that a comment written in
	// Vietnamese or German reads as prose just as an English one does.
	plainWord = regexp.MustCompile(`^\p{L}[\p{L}'-]*[,.]?$`)
)

// checkCommentedCode reports a run of comment lines that reads as shell code
// rather than as an explanation. It leaves alone the file header, shdoc
// blocks and their @example, directives, separators, a block that says why
// with a TODO, and a run that follows a sentence not yet ended, or one ending
// in `:` that introduces it as an example.
func checkCommentedCode(f *File, r *Reporter) {
	headerLast := len(fileHeader(f))
	if f.Shebang != "" && strings.HasPrefix(strings.TrimSpace(f.Line(1)), "#!") {
		headerLast++
	}
	var lines []int
	text := map[int]string{}
	eachComment(f, func(comment syntax.Comment) {
		line := int(comment.Pos().Line())
		if line <= headerLast || !strings.HasPrefix(strings.TrimSpace(f.Line(line)), "#") {
			return
		}
		if _, seen := text[line]; !seen {
			lines = append(lines, line)
		}
		text[line] = strings.TrimSpace(comment.Text)
	})
	sort.Ints(lines)
	for start := 0; start < len(lines); {
		end := start + 1
		for end < len(lines) && lines[end] == lines[end-1]+1 {
			end++
		}
		checkCommentBlock(r, lines[start:end], text)
		start = end
	}
}

// checkCommentBlock checks one block of comment lines that follow each other.
func checkCommentBlock(r *Reporter, lines []int, text map[int]string) {
	for _, line := range lines {
		body := text[line]
		if strings.HasPrefix(body, "@") || strings.HasPrefix(body, "#") || strings.Contains(body, "TODO") || strings.Contains(body, "FIXME") {
			return
		}
	}
	introduced := false
	for i := 0; i < len(lines); {
		if !codeLine(text[lines[i]]) {
			// A run after a sentence that has not ended, or that ends in `:`,
			// is the rest of that sentence or the example it introduces.
			body := text[lines[i]]
			introduced = body != "" && !strings.ContainsAny(body[len(body)-1:], ".!?")
			i++
			continue
		}
		j := i
		signal := false
		var run []string
		for ; j < len(lines) && codeLine(text[lines[j]]); j++ {
			run = append(run, text[lines[j]])
			signal = signal || strongCode(text[lines[j]])
		}
		if signal && !introduced && parsesAsShell(strings.Join(run, "\n")) {
			r.AtLine(lines[i], "this comment is code switched off; delete it, git keeps the old version, or say with a TODO why and when it comes back")
		}
		introduced = false
		i = j
	}
}

// codeLine reports whether a comment line could be a line of code: it is not
// a sentence, a directive or the label of an example.
func codeLine(body string) bool {
	if body == "" || strings.Contains(body, "`") || strings.HasPrefix(body, "shellcheck ") || strings.HasPrefix(body, "dyshellint") {
		return false
	}
	if strings.ContainsAny(body[len(body)-1:], ".:?!,") {
		return false
	}
	if sentenceStart.MatchString(body) || strings.Contains(body, "e.g.") || strings.Contains(body, "i.e.") {
		return false
	}
	words := strings.Fields(body)
	plain := 0
	for _, word := range words {
		// `label:` is prose: a word of code rarely ends with a colon.
		if strings.HasSuffix(word, ":") {
			return false
		}
		if plainWord.MatchString(word) {
			plain++
		}
	}
	// Three ordinary words or more, and mostly ordinary words: a sentence.
	return !(plain >= 3 && plain*10 >= len(words)*6)
}

// strongCode reports whether a line carries syntax that prose rarely does.
func strongCode(body string) bool {
	return codeSignal.MatchString(body) || codeKeyword[strings.Fields(body)[0]]
}

// parsesAsShell reports whether text is complete shell code.
func parsesAsShell(text string) bool {
	_, err := syntax.NewParser(syntax.Variant(syntax.LangBash)).Parse(strings.NewReader(text), "")
	return err == nil
}
