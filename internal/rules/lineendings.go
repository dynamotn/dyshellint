package rules

import (
	"bytes"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionLineEndings = "Formatting > Line Endings and Encoding"

func init() {
	register(Rule{
		Code:     "BSG140",
		Section:  sectionLineEndings,
		Severity: lint.SeverityError,
		Doc:      "Save shell files with LF line endings, in UTF-8 without a byte order mark",
		Check:    checkLineEndings,
		Bats:     true,
	})
}

// utf8BOM is the byte order mark some Windows editors put at the start of a
// UTF-8 file.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// checkLineEndings reports a byte order mark and the first CRLF line ending.
// The kernel reads the BOM as part of the shebang, so the script does not
// start; a CR stays at the end of every word, so `then\r` or a path ending in
// `\r` fails with a message that shows nothing wrong. One finding for the
// CRLF is enough: the whole file was saved that way.
func checkLineEndings(f *File, r *Reporter) {
	if bytes.HasPrefix(f.Src, utf8BOM) {
		r.AtLine(1, "the file starts with a UTF-8 byte order mark, which hides the shebang from the kernel; save it as UTF-8 without BOM")
	}
	if at := bytes.Index(f.Src, []byte("\r\n")); at >= 0 {
		line := bytes.Count(f.Src[:at], []byte("\n")) + 1
		r.AtLine(line, "the file has CRLF line endings, and Bash keeps the CR as part of each last word; convert it to LF, and add `*.sh text eol=lf` to `.gitattributes`")
	}
}
